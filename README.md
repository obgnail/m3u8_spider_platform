# m3u8 spider platform

<p align="center">
    <img src="assets/nijika.jpg" width="250" height="250">
</p>


## 简介

m3u8 系列网站爬虫，几乎可以通杀所有的 m3u8 的中小网站。


## 技术栈

为了通用性，选择了 `selenium` + `mitmproxy`。

- 隐蔽性强。
- 只写核心代码。直接从 API 层面处理。省的分析乱七八糟的 HTML。
- 解耦。selenium 负责爬取，mitmproxy 负责隐蔽和下载。
- selenium 保证了通用。小网站、盗版网站几乎都是盗链。前端反扒能力很低，几乎没有 selenium 特征识别。使用 selenium 可以轻松绕过。
- mitmproxy 保证了灵活。某些比较有心的网站还是会做 selenium 特征识别，此时只要使用 mitmproxy 拦截并废掉那些 JS 代码即可。


## 注意

存在一些特别小心（或者辣鸡后端）的网站，返回的 m3u8 媒体列表会有部分切片资源失效，需要更新 m3u8 文件后重复请求。这也是很多基于浏览器插件的嗅探软件（如 IDM）下载视频后文件内容缺失的原因。

对此，我加了循环校验功能，会循环直到所有切片都下载完成。


## License

MIT

> 为了使用 MIT 协议，没有使用 ffmpeg。



