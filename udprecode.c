
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <time.h>
#include <signal.h>
#include <sys/socket.h>
#include <sys/types.h>
#include <sys/wait.h>
#include <netinet/in.h>
#include <arpa/inet.h>
#include <pthread.h>
#include <errno.h>
#include <stdatomic.h>

#define MAX_THREADS 256
#define MAX_PACKET_SIZE 4096
#define STATS_INTERVAL 5

typedef struct {
    char ip[64];
    int port;
    int duration;
    int packet_size;
    int sockets_per_thread;
    int threads;
    atomic_llong *total_packets;
    atomic_llong *total_bytes;
} thread_config_t;

static volatile sig_atomic_t running = 1;

void sig_handler(int signum) {
    (void)signum;
    running = 0;
}

static inline uint64_t rdtsc() {
    uint32_t lo, hi;
    __asm__ __volatile__("rdtsc" : "=a" (lo), "=d" (hi));
    return ((uint64_t)hi << 32) | lo;
}

void* udp_thread(void *arg) {
    thread_config_t *cfg = (thread_config_t*)arg;
    struct sockaddr_in target_addr = {
        .sin_family = AF_INET,
        .sin_port = htons(cfg->port),
        .sin_addr.s_addr = inet_addr(cfg->ip)
    };

    // Allocate payload
    char *payload = aligned_alloc(64, cfg->packet_size);
    if (!payload) return NULL;
    
    memset(payload, 0x42, cfg->packet_size);

    // Create sockets
    int *sockets = malloc(cfg->sockets_per_thread * sizeof(int));
    if (!sockets) {
        free(payload);
        return NULL;
    }

    for (int i = 0; i < cfg->sockets_per_thread; i++) {
        sockets[i] = socket(AF_INET, SOCK_DGRAM, 0);
        if (sockets[i] < 0) continue;
        
        int sndbuf = 4 * 1024 * 1024; // 4MB buffer
        setsockopt(sockets[i], SOL_SOCKET, SO_SNDBUF, &sndbuf, sizeof(sndbuf));
        
        int reuse = 1;
        setsockopt(sockets[i], SOL_SOCKET, SO_REUSEADDR, &reuse, sizeof(reuse));
    }

    // Statistics
    uint64_t packets_sent = 0;
    uint64_t bytes_sent = 0;
    time_t last_stats = time(NULL);

    // Main sending loop
    while (running) {
        for (int i = 0; i < cfg->sockets_per_thread; i++) {
            if (sockets[i] < 0) continue;
            
            ssize_t sent = sendto(sockets[i], payload, cfg->packet_size, 0,
                                (struct sockaddr*)&target_addr, sizeof(target_addr));
            if (sent > 0) {
                packets_sent++;
                bytes_sent += sent;
            }
        }

        // Update global counters
        atomic_fetch_add(cfg->total_packets, packets_sent);
        atomic_fetch_add(cfg->total_bytes, bytes_sent);
        packets_sent = bytes_sent = 0;

        // Print stats every 5 seconds
        time_t now = time(NULL);
        if (now - last_stats >= STATS_INTERVAL) {
            printf("[Thread %ld] Packets/sec: %lld, Bandwidth: %.2f Mbps\n",
                   pthread_self() % 1000,
                   (long long)(*cfg->total_packets / STATS_INTERVAL),
                   (*cfg->total_bytes * 8.0) / (STATS_INTERVAL * 1000 * 1000));
            atomic_store(cfg->total_packets, 0);
            atomic_store(cfg->total_bytes, 0);
            last_stats = now;
        }
    }

    // Cleanup
    for (int i = 0; i < cfg->sockets_per_thread; i++) {
        if (sockets[i] >= 0) close(sockets[i]);
    }
    free(sockets);
    free(payload);
    return NULL;
}

int main(int argc, char *argv[]) {
    if (argc < 6) {
        fprintf(stderr, "Usage: %s <ip> <port> <duration> <threads> <sockets_per_thread> [packet_size=%d]\n",
                argv[0], MAX_PACKET_SIZE);
        return 1;
    }

    signal(SIGINT, sig_handler);
    signal(SIGTERM, sig_handler);

    thread_config_t cfg;
    strncpy(cfg.ip, argv[1], sizeof(cfg.ip));
    cfg.port = atoi(argv[2]);
    cfg.duration = atoi(argv[3]);
    cfg.threads = atoi(argv[4]);
    cfg.sockets_per_thread = atoi(argv[5]);
    cfg.packet_size = (argc > 6) ? atoi(argv[6]) : MAX_PACKET_SIZE;

    if (cfg.threads > MAX_THREADS) {
        fprintf(stderr, "Max threads: %d\n", MAX_THREADS);
        return 1;
    }
    if (cfg.packet_size > MAX_PACKET_SIZE) {
        fprintf(stderr, "Max packet size: %d\n", MAX_PACKET_SIZE);
        return 1;
    }

    printf("Starting UDP load test...\n");
    printf("Target: %s:%d, Duration: %ds, Threads: %d, Sockets/Thread: %d, Packet Size: %d\n",
           cfg.ip, cfg.port, cfg.duration, cfg.threads, cfg.sockets_per_thread, cfg.packet_size);

    atomic_llong total_packets = 0;
    atomic_llong total_bytes = 0;
    cfg.total_packets = &total_packets;
    cfg.total_bytes = &total_bytes;

    pthread_t threads[MAX_THREADS];
    for (int i = 0; i < cfg.threads; i++) {
        if (pthread_create(&threads[i], NULL, udp_thread, &cfg) != 0) {
            perror("pthread_create");
            return 1;
        }
    }

    // Wait for duration
    sleep(cfg.duration);
    running = 0;

    // Wait for threads
    for (int i = 0; i < cfg.threads; i++) {
        pthread_join(threads[i], NULL);
    }

    printf("\nTest completed!\n");
    printf("Total packets sent: %lld\n", (long long)total_packets);
    printf("Total bytes sent: %lld\n", (long long)total_bytes);
    printf("Average bandwidth: %.2f Mbps\n", (total_bytes * 8.0) / (cfg.duration * 1000 * 1000));

    return 0;
}