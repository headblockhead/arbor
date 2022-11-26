import json
from bs4 import BeautifulSoup
from html.parser import HTMLParser
import requests
import datetime
from PIL import Image


def get_attendance(kpis):

    html = json.loads(kpis)[
        "items"][0]["fields"]["html"]["value"]

    htmlsoup = BeautifulSoup(
        html, 'html.parser')

    for link in htmlsoup.find_all('div', class_='mis-htmlpanel-measure-value'):
        attendance = link.get_text()

    return attendance


def get_points(kpis):
    html = json.loads(kpis)[
        "items"][1]["fields"]["html"]["value"]
    htmlsoup = BeautifulSoup(
        html, 'html.parser')
    for link in htmlsoup.find_all('div', class_='mis-htmlpanel-measure-value'):
        points = link.get_text()
    return points


def get_name(login_reciv):
    loginjson = json.loads(login_reciv.text)
    name = loginjson["items"][0]["display_name"]
    return name


def get_profile_img(headers, arborURL):
    url = arborURL + 'students/home-ui/dashboard'
    response = requests.get(url, headers=headers)

    responsejson = json.loads(response.text)
    content = responsejson["content"][0]["content"][0]["props"]["picture"]

    imgurl = content.split("/circle/1")[0]
    profileimg = requests.get(imgurl, headers=headers)

    file = open("/temp/profile.bmp", "wb")
    file.write(profileimg.content)
    file.close()

    img = Image.open("/temp/profile.bmp")
    img = img.convert("1").crop((30, 30, 160, 190))

    return img


def get_data(headers, arborURL):
    formattedDate = get_date()

    url = arborURL + 'calendar-entry/list-static/format/json/'

    rawData = '{"action_params":{"view":"day","startDate":"' + formattedDate + '","endDate":"' + \
        formattedDate + \
        '","filters":[{"field_name":"object","value":{"_objectTypeId":1,"_objectId":4378}}]}}'
    response = requests.post(url, headers=headers, data=rawData)

    jsonData = json.loads(response.text)

    jsonDataResponseValue = jsonData["items"][0]["fields"]["response"]["value"]
    start = jsonDataResponseValue["currentView"]["start"]
    pages = jsonDataResponseValue["pages"]

    for page in pages:
        if (page["start"] != start):
            continue
        html = page["html"]
        soup = BeautifulSoup(page["html"], 'html.parser')
        subjects = []
        times_loactions = []
        i = 0
        for link in soup.find_all('b'):
            i += 1
            if (i % 2 == 0):
                subjects.append(link.get_text())
        i = 0
        for link in soup.find_all('span'):
            i += 1
            if (i % 2 == 1):
                times_loactions.append(link.get_text())
        final_string = ""
        for i in range(len(subjects)):
            final_string = final_string + \
                (subjects[i] + " | " + times_loactions[i]) + "\n"
        return final_string


def get_week(headers, arborURL):
    formatted_date = get_date()

    url = arborURL + 'calendar-entry/list-static/format/json/'
    raw_data = '{"action_params":{"view":"day","startDate":"' + formatted_date + '","endDate":"' + \
        formatted_date + \
        '","filters":[{"field_name":"object","value":{"_objectTypeId":1,"_objectId":4378}}]}}'
    resp = requests.post(url, headers=headers, data=raw_data)

    # some JSON:
    x = resp.text

    # parse x:
    y = json.loads(x)

    items = y["items"]
    items_zero = items[0]
    fields = items_zero["fields"]
    response = fields["response"]
    value = response["value"]
    currentView = value["currentView"]
    start = currentView["start"]
    pages = value["pages"]

    for page in pages:
        if (not "Week" in page["title"]):
            continue
        else:
            return page["title"].split("(Week ")[1].split(")")[0]


def get_date():
    x = datetime.datetime.now()
    formatted_date = x.strftime("%Y-%m-%d")
    return formatted_date


def get_kpis(headers, arborURL):
    url = arborURL + 'students/student/kpis/'

    response = requests.get(
        url, headers=headers)
    return response.text


def get_headers(login_reciv):
    loginjson = json.loads(login_reciv.text)
    session_id = loginjson["items"][0]["session_id"]
    return {'cookie': '__stripe_mid=0b22f1a8-fb07-479d-a071-dc37b0c3ec8ab82942; __stripe_sid=b8539c79-5620-42b7-b7ff-7d9f47a3bb0030ad59; mis=' + session_id}


def login(url, user, password):
    raw_data_login = '{"items":[{"username":"' + user + '","password":"' + password + '"}]}'
    login_reciv = requests.post(url, data=raw_data_login)
    return login_reciv
