# 用 Go 语言写焦虑发生器并发布到 Rum 上·第一篇

初学 Go 语言是在去年 11 月 20 日。到今年 2022 年的 2 月 3 日，我写了一个小 bot 连续运行成功，发布到了 github 上开源。花了两个多月时间。我感觉效率还是不错的。
于是我这就来记录这段学习经历：
目的一是让同样正在跨编程入门这道薛定谔之忽高忽低门槛的小白同学一些参考，更顺利的入门编程；
目的二是回顾并巩固自己过去的学习，为下一步继续学习打好基础；
目的三是让更多的人能够对 Rum 这个新的东西感兴趣。
希望能成功达成目的：

## Go 语言基础入门
学 Go 的出发点是因为大佬们都用 Go 语言来写链上应用。看着他们在群里交流的非常欢乐，文字我都认识，但就不懂他们在讲什么。
这种感觉太难受，就像小矮人身在片场却眼巴巴看高等精灵们讲精灵语，脸上还是挂着围笑假装在参与对话。

然后呢，因为自己喜欢游戏，找了本书叫：[Pac Go: A Pac Man clone written in Go
](https://xue.cn/hub/app/books/236)
就是用 Go 语言写吃豆人游戏。好，从这本书开始入门 Go。
并不是因为作者是美丽的女程序员我才选择的这本书。
![[img/author_danicat.png]]

>**并不是**

```P.S. 这本书的链接是链到  [xue.cn](https://xue.cn)  的，是一个可以一边阅读编程教程一边在当前书页运行代码的学习网站。```

一边做实例，一边学编程，也非常符合我自己的学习理念。很快我从  [xue.cn](https://xue.cn) 转移到了 github 去学习，上面有 Pac Go 的开源代码。自己 clone 了一份到电脑上，然后通过读 readme 继续学习。

老实说 Pac Go 的前 5 章我认真跟着学了，后面就没有认真学，而是略读之后，感觉基础语法已经掌握差不多了，就开始用 Go 做自己的项目。
因为新的一年快要到了，如何让人即便是年初，也要焦虑起来呢？想到了在 Twitter 上见过的 year progress。把人习以为常的日期，转变成百分比的进度条，会发现，时间怎么这么不经用？这么不经意的流逝掉了？这会让人产生巨大的焦虑感。于是我打算做一个这样的 bot 并到 Rum 上去运行。

## Go 语言发送 HTTP request
理性的角度出发我应该规划一下这个程序的各个功能零件，以及工作流程，然后从生成进度条这一步开始，最后再做发布到 Rum 的 HTTP 请求。但我是小白啊，我野路子啊！我有搞着玩的特权啊！
先试着用 Go 写 HTTP 的 Post 请求，发布到 Rum 上看看再说。

导入Go 语言的 HTTP 包：
``` Go
import "net/http"
```

有了这个包，就可以调用 Go 语言的 HTTP 方法了，这里我是随便 google 了一下，了解到 HTTP 到底是个什么东西，HTTP 请求又是怎么发送的。
在此建议像我一样的新手小白也自己去研究 HTTP，这个并不难。本文不再花篇幅来讲 HTTP。
当然深入研究也会花大量时间，这里用不到那么深入的知识，单独看看 Post 和 Get 两种最常用的 HTTP 请求就好了。

简单理解了 HTTP 请求，回来继续写自己的代码。

建立一个 client 用来发送 request。代码是：
``` Go
tr := &http.Transport{
	TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
}
client := &http.Client{Transport: tr}
```
上面这一段代码简单的解释是这样的：
变量 client 是一个地址，指向了 http 包内的 struct，名为 Client。而这个 Client 里的一个值，也是一个 struct，名为 Transport。将 Transport 里的一个 TLSClientConfig 写入一个 tls 设置，把 InsecureSkipVerify 设置为 true。
这里的 tls 是把 HTTP 变成 HTTPS 的一层协议。我们这个设置是为了跳过一个 tls 的验证。因为这个 HTTP 请求的地址就在本地，我们可以不用进行验证。

client 建立之后我们需要用 client 来发送请求。这段代码是：
``` Go
req, err := http.NewRequest("POST", "https://127.0.0.1:[Rum节点的端口号]/api/v1/group/content", body)
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()
```

变量 req 就是我们要发送请求的一个实例了。通过 http.NewRequest 来建立，有三个参数，分别是：
"POST"，表示我们的请求是 POST 方法；
第二个参数是 URL，端口号可以在 Rum 客户端的“节点与网络”菜单中的“节点参数”中找到；
![[img/portnumber.png]]

第三个 body 变量是要 Post 给 Rum 的具体内容。

接下来设置一个 header，把 header 设置为一个 json 的内容。这是因为 Rum 需要我们发送 json 内容。
最后就通过 client.Do(req) 来执行我们设置好的一切，并将请求到的返回值赋值给 resp 变量。这样就通过 Go 完成了一个完整的 HTTP request。

我把这个完整的 HTTP request 提供如下，整个 request 写成了一个叫 postToRum 的函数，请注意函数里面定义的叫 Payload 的 struct 数据结构是按照 quorum 的格式要求来声明的，内容格式可以自定义的是标题，正文，然后目标种子网络的ID，其他的不用修改：
``` Go
func postToRum(title string, content string, group string, url string) {
	type Object struct {
		Type    string `json:"type"`
		Content string `json:"content"`
		Name    string `json:"name"`
	}
	type Target struct {
		ID   string `json:"id"`
		Type string `json:"type"`
	}
	type Payload struct {//按照 quorum 要求的数据结构进行声明
		Type   string `json:"type"`
		Object Object `json:"object"`
		Target Target `json:"target"`
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	data := Payload{
		Type: "Add",
		Object: Object{
			Type:    "Note",
			Content: content,
			Name:    title,
		},
		Target: Target{
			ID:   group,
			Type: "Group",
		},
	}

	payloadBytes, err := json.Marshal(data)
	if err != nil {
		panic(err) // handle err
	}

	fmt.Println(string(payloadBytes))

	body := bytes.NewReader(payloadBytes)

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	received, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(received))
}
```

函数的四个参数分别是 title 表示标题，content 表示内容，group 用于指定要发布内容的种子网络 ID，最后 url 是要发 POST 请求的目标 url，这里的地址是根据 Rum 的 api 要求来的，读者感兴趣可以自己在 Rum 的 github 主页去看看，这里的话可以直接用我提供的地址。

值得注意的是，请求的内容主体：body 变量，经过了两次加工：
最初是一个 struct，这个 struct 要符合 Rum 的格式要求，我们取名叫 Payload。
然后用 Payload 来创建一个叫 data 的实例，给 data 填入了具体的内容。
再接下来，用 json.Marshal(data) 方法，把 data 解析成了 json 格式，并赋值给变量 payloadBytes。
最后再把 payloadBytes 通过 bytes 包的 bytes.NewReader(payloadBytes) 方法，转变成了能够通过 HTTP POST 方法发送给 Rum 的字符。

既然写好了这个函数，我也迫不及待的往 Rum 上发了一个 Hello Rum 的消息。
于是在 mian 函数里写下如下代码：
``` Go
func main() {
	url := "https://127.0.0.1:[端口号]/api/v1/group/content"
	postToRum("Hello Rum", "Hello Rum", "[目标种子网络的ID]", url)

}
```
目标种子网络的 ID 可以在种子网络的详情处获取到，比如“Go语言学习小组”的 ID 是
>fe2842cb-db6b-4e8a-b007-e83e5603131c

![[img/groupID.png]]

我们填入 ID 就可以往“Go语言学习小组”发送 Hello Rum 了。
以上代码的片段忽略掉了一些 Go 语言的一些前置语句，比如包管理的 package 语句，比如引入依赖的包的 import 语句。这里我把代码完整的提供到了 github 仓库里。本系列文章的第一步骤放在了 Step0 文件夹里（因为我们程序员要习惯从零开始）：

https://github.com/hawken-im/yearprogress/tree/main/Step0

clone 完整代码或者复制粘贴也行，之后在 Step0 目录下执行：
```
go run main.go
```

就能看到结果了。

非常欢迎读者们发送 Hello Rum 到“Go语言学习小组”，小组的种子提供于此：
```
{
  "genesis_block": {
    "BlockId": "7016d356-b42f-421c-a086-094e1f35dbeb",
    "GroupId": "fe2842cb-db6b-4e8a-b007-e83e5603131c",
    "ProducerPubKey": "CAISIQOU1kDjMc3cCZRKV/r2bU/IUPukEcdFkIqkFe3Gbqfy+w==",
    "Hash": "v+kfzMMuwNgb2h1PUAktBk1K9DZbN9pEdcfg2rG1Zys=",
    "Signature": "MEUCIAZ8A4fgP5TWjXZoAe47qqfktrMrP1/2MMsOM5QsaFiQAiEAn8i8SzpdbGd4wlbbtk6Dws32Ea6aBWtcam+VdUzeHBg=",
    "TimeStamp": "1637338394235167000"
  },
  "group_id": "fe2842cb-db6b-4e8a-b007-e83e5603131c",
  "group_name": "GO语言学习小组",
  "owner_pubkey": "CAISIQOU1kDjMc3cCZRKV/r2bU/IUPukEcdFkIqkFe3Gbqfy+w==",
  "consensus_type": "poa",
  "encryption_type": "public",
  "cipher_key": "835360cc49a5faf385b906b8fd1fb16f31a73c652c65398513070c27a3920550",
  "app_key": "group_post",
  "signature": "304502204baef7f83e01af403791a96024413deb59ecec7b92f9ae2c18377917e127e6c1022100a4529dc2542aa3f6dc9afd8d14d8bfbcbb3ac33a3a32ca993805a13a77942efe"
}
```
