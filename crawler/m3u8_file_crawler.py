import time

from selenium import webdriver
from selenium.webdriver.chrome.options import Options

line = 0


def wait():
    global line
    for _ in range(0, 16):
        with open(r'd:\volume\mitmproxy\scripts\m3u8_file.txt', 'r') as f:
            c = f.read().count('\n')
            if c > line:
                line = c
                return
        time.sleep(3)
    raise "timeout"


def get_webdriver(path, proxy='', headless=False):
    options = Options()
    options.add_experimental_option('excludeSwitches', ['enable-automation'])
    options.add_argument('--disable-blink-features=AutomationControlled')
    if headless:
        options.add_argument("--headless")
    if proxy:
        options.add_argument('--proxy-server=%s' % proxy)
    driver = webdriver.Chrome(executable_path=(path), options=options)
    driver.maximize_window()
    return driver


def play_video(driver, episode):
    for num in range(1, episode + 1):
        wait()
        print(f'#{num} start...')
        driver.execute_script(f'play(0,{num - 1});')


def crawl():
    chrome_driver_path = r'd:\tmp\chromedriver.exe'
    url = 'https://www.nunuyy5.org/dianshiju/5249.html'
    proxy = '127.0.0.1:8080'
    episode = 84

    driver = get_webdriver(chrome_driver_path, proxy=proxy)
    try:
        driver.get(url)
        play_video(driver, episode)
        # time.sleep(10000)
    finally:
        driver.close()


if __name__ == '__main__':
    crawl()
