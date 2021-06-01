#!/bin/sh
# Install script for termux

install_pkg(){
	 pkg i python termux-api vim-python
}

install_requirements(){
    pip install -r requirements.txt
}

install_pkg && install_requirements 
