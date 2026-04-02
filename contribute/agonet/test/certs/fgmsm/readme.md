证书拷贝于
https://github.com/tjfoc/gmsm/tree/master/gmtls/websvr/certs


国密 SM2 双证书体系的典型文件集合，分为CA 根证书 / 密钥和实体证书 / 密钥两类，核心遵循《GMT 0024-2015 证书应用技术规范》。

## 根 CA 相关文件（信任根）
`SM2_CA.cer`：SM2 算法的根 CA 证书（公钥文件），用于验证所有由该 CA 签发的实体证书合法性。
`SM2_CA_KEY.pem`：根 CA 的私钥文件，用于给实体证书签名（非常敏感，需严格保管）。

> 作用：作为信任链的顶端，所有终端实体证书都由这对 CA 密钥签发。


文件	用途	对应关系

- sm2_sign_cert.cer:	签名证书（公钥）：用于数字签名、身份认证、防抵赖	配对 sm2_sign_key.pem（签名私钥）
- sm2_sign_key.pem:	签名私钥：用于生成数字签名，必须保密	对应 sm2_sign_cert.cer
- sm2_enc_cert.cer:	加密证书（公钥）：用于数据加密、密钥协商	配对 sm2_enc_key.pem（加密私钥）
- sm2_enc_key.pem:	加密私钥：用于解密被加密的数据，必须保密	对应 sm2_enc_cert.cer
- sm2_auth_cert.cer:	认证证书（公钥）：部分场景下用于身份认证（可与签名证书合并，也可单独拆分）	配对 sm2_auth_key.pem（认证私钥）
- sm2_auth_key.pem:	认证私钥：用于身份认证场景	对应 sm2_auth_cert.cer