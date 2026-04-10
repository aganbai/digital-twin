#!/usr/bin/env python3
# -*- coding: utf-8 -*-
import socket
import urllib.request

def check_port(host, port):
    """检查服务是否运行"""
    try:
        sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        sock.settimeout(3)
        result = sock.connect_ex((host, port))
        sock.close()
        print(f"Socket check result for {host}:{port}: {result}")
        return result == 0
    except Exception as e:
        print(f"Socket check error: {e}")
        return False

def check_http(host, port):
    """HTTP检查"""
    try:
        url = f"http://{host}:{port}/"
        req = urllib.request.Request(url, method='GET')
        req.add_header('User-Agent', 'Mozilla/5.0')
        response = urllib.request.urlopen(req, timeout=5)
        print(f"HTTP check for {host}:{port}: status={response.status}")
        return response.status == 200
    except Exception as e:
        print(f"HTTP check error: {e}")
        return False

print("Checking port 5174...")
port_ok = check_port("localhost", 5174)
print(f"Port open: {port_ok}")

print("\nChecking via HTTP...")
http_ok = check_http("localhost", 5174)
print(f"HTTP ok: {http_ok}")
