#!/usr/bin/python
import subprocess
import json

# Get Mobile last msg for otp Checking  
def get_msg():
    msg = {}

    msg = subprocess.Popen(
                            'termux-sms-list -l 1',
                            stdin=subprocess.DEVNULL,
                            stdout=subprocess.PIPE,
                            stderr=subprocess.PIPE,shell=True).communicate()[0].decode('utf-8')
    try:
        msg = json.loads(msg)[0]
        return msg
    except KeyError:
        raise Exception("Install Termux:API 0.31 Version for AutoMode")

if __name__== "__main__":
    txt = get_msg()
    print(txt)
