#!/usr/bin/python
# -*- coding:utf-8 -*-

# import
from PIL import ImageDraw
from PIL import ImageFont
from PIL import ImageColor
from PIL import Image

import requests
import urllib
import re
import os

# sudo python3 -m pip install requests


def Get_Html(Url):  # Timeout retry (failed if not successful 3 times)
    i = 0
    while i < 10:
        try:
            html = requests.get(Url, timeout=5)
            return html
        except requests.exceptions.RequestException as e:
            logging.debug(e)
            i += 1
    if(i == 3):
        logging.critical("Network connection failed")
        sys.exit(0)


def Get_info(source, path):
    # for i in range len(path):
    if(len(path) > 1):
        return Get_info(source[path[0]], path[1:])
    else:
        return source[path[0]]


def Get_PNG(Url, Name, Width=400, Heigth=300):
    import requests
    r = Get_Html(Url)
    with open(Name, 'wb') as f:
        for chunk in r.iter_content(chunk_size=32):
            f.write(chunk)
    f.close()

    # Read the file and convert it to a bitmap
    # Increase portability
    PNG = Image.open(Name).convert('LA')
    gray = PNG.convert('L')
    # Turn into bitmap and color flip
    bw = gray.point(lambda x: 0 if x < 200 else 255, '1')
    bw = bw.resize((Width, Heigth), Image.ANTIALIAS)
    bw.save(Name)  # save Picture
    return bw


class Spider_baidu_image():
    def __init__(self, keyword='黑白', paginator=1):
        self.url = 'http://image.baidu.com/search/acjson?'
        self.headers = {
            'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/69.0.\
            3497.81 Safari/537.36'}
        self.headers_image = {
            'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/69.0.\
            3497.81 Safari/537.36', 'Referer': 'http://image.baidu.com/search/index?tn=baiduimage&ipn=r&ct=201326592&cl=2&lm=-1&st=-1&fm=result&fr=&sf=1&fmq=1557124645631_R&pv=&ic=&nc=1&z=&hd=1&latest=0&copyright=0&se=1&showtab=0&fb=0&width=&height=&face=0&istype=2&ie=utf-8&sid=&word=%E8%83%A1%E6%AD%8C'}
        # must encoding utf-8
        self.keyword = keyword
        #print ( chardet . detect ( str . encode ( self.keyword ) ) )

        self.paginator = paginator
        # print(type(self.keyword),self.paginator)
        # exit()

    def get_param(self):
        """
        get url parameter and return
        获取url请求的参数，存入列表并返回 
        """
        keyword = urllib.parse.quote(self.keyword)
        params = []
        params.append('tn=resultjson_com&ipn=rj&ct=201326592&is=&fp=result&queryWord={}&cl=2&lm=-1&ie=utf-8&oe=utf-8&adpicid=&st=-1&z=&ic=&hd=1&latest=0&copyright=0&word={}&s=&se=&tab=&width=&height=&face=0&istype=2&qc=&nc=1&fr=&expermode=&force=&cg=star&pn={}&rn=30&gsm=78&1557125391211='.format(keyword, keyword, 30 * self.paginator))
        return params

    def get_urls(self, params):
        """
        由url参数返回各个url拼接后的响应，存入列表并返回
        get url parameter response and return
        """
        urls = []
        for i in params:
            urls.append(self.url + i)
        return urls

    def get_image_url(self, urls):
        """
        获取图片url，存入列表并返回
        get image url and return
        """
        image_url = []
        for url in urls:
            json_data = requests.get(url, headers=self.headers)
            # print(json_data.text)
            json_data = json_data.json()
            json_data = json_data.get('data')
            for i in json_data:
                if i:
                    image_url.append(i.get('thumbURL'))
        return image_url

    def get_image(self, image_url, Width, Heigth):
        """
        创建image文件夹，将图片存在其中
        Create an Image folder and store images in it
        """
        cwd = os.getcwd()
        file_name = os.path.join(cwd, "image")
        #file_name = "image"
        if not os.path.exists(file_name):
            os.mkdir(file_name)

        file_name = file_name + "/"
        for index, url in enumerate(image_url, start=1):
            image_name = file_name + '{}.jpg'.format(index % 30)
            with open(image_name, 'wb') as f:
                f.write(requests.get(url, headers=self.headers_image).content)

            IMG = Image.open(image_name).convert('LA')
            GRAY = IMG.convert('L')
            # Turn into bitmap and color flip
            bw = GRAY.point(lambda x: 0 if x < 200 else 255, '1')
            bw = bw.resize((Width, Heigth), Image.ANTIALIAS)
            bw.save(image_name)  # save Picture
            # return bw

    def __call__(self, Width=400, Heigth=300):
        params = self.get_param()
        urls = self.get_urls(params)
        image_url = self.get_image_url(urls)
        return self.get_image(image_url, Width, Heigth)
