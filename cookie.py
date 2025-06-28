import os

def check_sqlite():
    if os.system("command -v sqlite3") != 0:
        print("❌ SQLite chưa được cài đặt, cài đặt ngay...")
        os.system("pkg install sqlite -y")
    else:
        print("✅ SQLite đã có sẵn!")

cookie_db = "~/storage/shared/Android/data/com.android.chrome/app_chrome/Default/Cookies"
cmd_all_cookies = f'sqlite3 {cookie_db} "SELECT host_key, name, value FROM cookies;"'
cmd_save_cookies = f'sqlite3 {cookie_db} "SELECT name, value FROM cookies;" > cookie.txt'

def get_cookie():
    print("📥 Đang lấy cookie từ Chrome...\n")    
    os.system(cmd_all_cookies)
    os.system(cmd_save_cookies)
    print("\n✅ Cookie đã được lưu vào cookie.txt!")

if __name__ == "__main__":
    check_sqlite()
    get_cookie()
    