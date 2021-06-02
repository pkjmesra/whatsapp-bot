# WhatsApp Bot ![Go](https://github.com/pkjmesra/whatsapp-bot/workflows/Go/badge.svg)
Whatsapp Bot for CoWIN vaccination (with automated CAPTCHA verification)

*This is absolutely for academic reference purposes only. Lagalities arising out of commercial or personal use is none of the concerns of the author(s) of this and/or any related repo*

![WhatsApp Icon](https://cdn.icon-icons.com/icons2/373/PNG/96/Whatsapp_37229.png)
![CoWIN Icon](https://prod-cdn.preprod.co-vin.in/assets/images/covid19logo.jpg)

## Install

```
go get github.com/pkjmesra/whatsapp-bot
```

## Getting Started on your PC/Mac

```go
$ go build && go install && mv $GOPATH/bin/whatsapp-bot ./whatsapp-bot

$ ./whatsapp-bot -i 30
```
- i gives the interval to ping and poll for available vaccination slots for a given set of parameters
- m can help setup a specific mobile number for which you'd like to book an appointment via whatsapp

When you run, your mobile/whatsapp will become the host for sending/receiving data with [91MobileNumber].

- On your mobile phone, launch whatsapp and send a "subscribe" message to yourself or ask your friend to send a "subscribe" message to your number. (The same number with which you've logged in into whatsapp-bot)

- Every API call is automated for CoWIN. 
- OTP is sent to your mobile <91MobileNumber>. Enter it in whatsapp when prompted or follow the instructions below to automatically have it entered.
- CAPTCHA will be sent on whatsapp itself. Enter CAPTCHA to proceed for appointment if not autodetected. In 95% of the cases, CAPTCHA will be autodetected and entered.

Entire transaction can be finished in : time taken for OTP + time taken for CAPTCHA (if not detected automatically).

## Getting Started on Android
  On your Android phone having the whatsapp running with the same number with which you logged in in the previous step, please do the following setup for OTP.

  By using Tremux you can run script and receive the notification on your phone.

  - ### Install Termux 

    - Install Termux App [FDroid](https://f-droid.org/en/packages/com.termux/).
    ##### Termux wiki suggest not to use playstore termux.

    
 - ### Installing Packages and Requirements
   - Step 1 : First update pkg
    
         pkg update

   - Step 2 : Install git

         pkg install git

   - Step 3 : Clone repo 

         git clone https://github.com/pkjmesra/whatsapp-bot.git
        
   - Step 4 : Open Cloned Folder
        
         cd whatsapp-bot

   - Step 5: run install.sh 
         
         bash install.sh

   - Step 6: run otp.py
         
         cd pkBot

         chmod +x otp.py

         python otp.py 91Your-Mobile-Number-Here-With-Which-You-Have-Launched-WhatsApp

When OTP is received, it will launch whatsapp on your mobile to send it to the number with which it was launched on the PC/Mac.

Now open Whatsapp on your Android and send type : *subscribe* and send it to the same number with which you have logged in on the whatsapp-bot on your Mac/PC.

Enjoy!

- Subscribe to getting notified
![](./images/1.jpg)


- Setting up your search parameters
![](./images/2.jpg)


- Saving your search parameters and confirming
![](./images/3.jpg)


- Receiving your registered beneficiaries
![](./images/4.jpg)


- Receiving the notification when the slots become available
![](./images/5.jpg)


- Trying to book the appointment and receive the confirmation or message if it could not be booked.
![](./images/6.jpg)


After launching, app will create a [WhatsApp connection](https://github.com/Rhymen/go-whatsapp). If you are not logged in, it will print a QR code in the terminal. Scan it with your phone and you are ready to go!

> Bot will remember the session (TempDir > pkBot > Session > WhatsappSession.gob) so there is no need to authenticate everytime.

