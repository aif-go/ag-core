#!/bin/bash

# 创建临时配置文件用于添加 SAN 扩展
cat > san.cnf << EOF
[req]
distinguished_name = req_distinguished_name
x509_extensions = v3_ca
req_extensions = v3_req
prompt = no

[req_distinguished_name]
CN = placeholder

[v3_ca]
subjectKeyIdentifier = hash
authorityKeyIdentifier = keyid:always,issuer
basicConstraints = critical, CA:true
keyUsage = critical, digitalSignature, cRLSign, keyCertSign

[v3_req]
subjectAltName = @alt_names

[alt_names]
DNS.1 = localhost
IP.1 = 127.0.0.1
EOF

# 生成 CA 密钥和证书（添加 SAN 扩展）
openssl genrsa -out ca.key 2048
openssl req -new -x509 -days 3650 -key ca.key -out ca.crt -subj "/CN=Test CA" -config san.cnf -extensions v3_ca

# 生成服务器密钥和证书请求（指定 SAN）
openssl genrsa -out server.key 2048
openssl req -new -key server.key -out server.csr -subj "/CN=localhost" -config san.cnf -extensions v3_req

# 使用 CA 签名服务器证书（添加 SAN 扩展）
openssl x509 -req -days 365 -in server.csr -CA ca.crt -CAkey ca.key -set_serial 01 -out server.crt -extfile san.cnf -extensions v3_req

# 生成客户端密钥和证书（客户端也添加 SAN）
openssl genrsa -out client.key 2048
openssl req -new -key client.key -out client.csr -subj "/CN=client" -config san.cnf -extensions v3_req
openssl x509 -req -days 365 -in client.csr -CA ca.crt -CAkey ca.key -set_serial 02 -out client.crt -extfile san.cnf -extensions v3_req
