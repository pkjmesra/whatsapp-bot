#!/usr/bin/python
import subprocess
import json
import sys
import re
import time
import datetime
import urllib.request

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
    otp = ""
    if sys.argv[1] == '':
        print('Invoke this with your mobile Number')
        exit(1)
    try:    
        curr_msg = get_msg()
        curr_msg_body = curr_msg.get("body")
        for i in reversed(range(30)):
            last_msg = get_msg()
            last_msg_body = last_msg.get("body",'')
            print(f'Waiting for OTP {i} sec')
            sys.stdout.write("\033[F")
            d1 = datetime.datetime.strptime(last_msg.get("received"), '%Y-%m-%d %H:%M:%S')
            now = datetime.datetime.now() # current date and time
            d2 = datetime.datetime.strptime(now.strftime("%Y-%m-%d %H:%M:%S"), '%Y-%m-%d %H:%M:%S')
            diff = (d2 - d1).total_seconds()
            if (curr_msg_body != last_msg_body and "cowin" in last_msg_body.lower()) or diff <= 180:
                otp = re.findall("(\d{6})",last_msg_body)[0]
                print(f'')
                sys.stdout.write("\033[F")
                break
            time.sleep(1)
    except (Exception,KeyboardInterrupt) as e:
        print(e)
    print(otp)
    urllib.request.urlopen('whatsapp://send?phone=' + sys.argv[1] + '&text='+otp)
