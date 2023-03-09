# DRatConfig
用于自动上传配置文件到ENS的工具
## 参数说明
```
-ens string
    ENS域名
-json string
    配置文件路径
-pk string
    私钥
-rpc string 
    RPC地址 (选填，默认 "https://goerli.infura.io/v3/9aa3d95b3bc440fa88ea12eaa4456161")
```
## 特别说明
- 存在网络拥堵的情况可以等待一段时间（通常北京时间下午会非常拥堵，第二天早上就好了），用TxHash去查询状态
- 有能力的用户可以自行修改该加密算法，当然DRat中的也要修改
## 免责声明
本工具仅用于学习交流，不得用于非法用途，否则后果自负