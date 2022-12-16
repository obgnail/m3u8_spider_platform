from mitmproxy import ctx


class Filter:
    '''屏蔽selenium检测'''

    def __init__(self):
        self.list = ['webdriver', '__driver_evaluate', '__webdriver_evaluate', '__selenium_evaluate',
                  '__fxdriver_evaluate', '__driver_unwrapped', '__webdriver_unwrapped',
                  '__selenium_unwrapped', '__fxdriver_unwrapped', '_Selenium_IDE_Recorder',
                  '_selenium', 'calledSelenium', '_WEBDRIVER_ELEM_CACHE', 'ChromeDriverw',
                  'driver-evaluate', 'webdriver-evaluate', 'selenium-evaluate', 'webdriverCommand',
                  'webdriver-evaluate-response', '__webdriverFunc', '__webdriver_script_fn',
                  '__$webdriverAsyncExecutor', '__lastWatirAlert', '__lastWatirConfirm',
                  '__lastWatirPrompt', '$chrome_asyncScriptInfo', '$cdc_asdjflasutopfhvcZLmcfl_']

    def response(self, flow):
        if '.js' in flow.request.url:
            for webdriver_key in self.list:
                ctx.log.info('Remove "{}" from {}.'.format(webdriver_key, flow.request.url))
                flow.response.text = flow.response.text.replace('"{}"'.format(webdriver_key), '"NO-SUCH-ATTR"')
            flow.response.text = flow.response.text.replace('t.webdriver', 'false')
            flow.response.text = flow.response.text.replace('ChromeDriver', '')


class M3U8Blocker():
    def __init__(self):
        self.dict = {}

    def request(self, flow):
        url = flow.request.url
        if 'b.baobuzz.com/m3u8' in url:
            if not self.dict.get(url):
                print('--- get m3u8 url:', url)
                with open('/tmp/m3u8_file.txt', 'a+') as f:
                    f.write(url + '\n')
                self.dict[url] = 1
            # flow.response = http.Response.make(404)

    def response(self, flow):
        '''废掉某些js'''
        if 'qrcode.js' in flow.request.url:
            flow.response.text = flow.response.text.replace('RS_BLOCK_TABLE', '')
            # flow.response = http.Response.make(404)


addons = [
    # Filter(),
    M3U8Blocker(),
]
