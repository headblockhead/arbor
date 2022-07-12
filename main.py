#!/usr/bin/python

import os
import sys

# Use our custom libraries to import the necessary modules (this needs to go above the import of the modules)
picdir = 'sources'
libdir = 'lib'
if os.path.exists(libdir):
    sys.path.append(libdir)

from PIL import ImageDraw
from PIL import ImageFont
from PIL import ImageColor
from PIL import Image

from waveshare_epd import waveshare_epd
from tcp_server import tcp_sver
import socketserver
import traceback
import threading
import logging
import struct
import signal
import fcntl
import time
import math
from datetime import date
from parsedata import *
from progressbar import *
from time import *

credentialsLocation = 'creds.json'
fontPath = 'fonts/ubmr.ttf'
# deviceSize is in inches, can be 4.2 or 2.7
deviceSize = 4.2

# Create a custom tcp server to serve the latest timetable to the e-paper client


class TimetableServer(tcp_sver.tcp_sver):
    def handle(self):
        try:
            self.client = self.request
            deviceID = self.Get_ID()

            print("Device connecting of ID:", deviceID)

            credsFile = open(credentialsLocation)
            credsJSON = json.load(credsFile)
            credentials = login(credsJSON['url'] + 'auth/login',
                                credsJSON['user'], credsJSON['pass'])
            arborURL = credsJSON['url']
            targetedDevice = credsJSON['device_id']
            devicePassword = credsJSON['device_password']
            credsFile.close()

            if (deviceID != targetedDevice):
                print("Device is not owned by me, ignoring")
                return

            # If device is locked, assume the password is as the user set
            self.unlock(devicePassword)

            # Setup the device with a certain size screen.
            epd = waveshare_epd.EPD(deviceSize)
            self.set_size(epd.width, epd.height)

            headers = get_headers(credentials)

            weekNumber = get_week(headers, arborURL)

            currentDay = date.today().strftime('%A')
            formattedDate = get_date()

            kpis = get_kpis(headers, arborURL)
            attendance = get_attendance(kpis)
            termPoints = get_points(kpis)

            fullName = get_name(credentials)

            calendarData = get_data(headers, arborURL)
            splitCalendarData = calendarData.split('\n')

            # Setup the fonts to be used
            # UBMR = Ubuntu Mono Regular (https://fonts.google.com/specimen/Ubuntu+Mono)

            fontTitle = ImageFont.truetype(fontPath, 40)
            fontSubtitle = ImageFont.truetype(fontPath, 20)
            fontEvents = ImageFont.truetype(fontPath, 12)

            # Create a new image to draw on of the size of the device's screen. Clear the frame with white
            virtualImage = Image.new('1', (epd.width, epd.height), 255)

            # Create a new drawing object to draw on the virtual image
            draw = ImageDraw.Draw(virtualImage)

            # Draw the text labels on the screen.
            draw.text((0, 0), 'Timetable', font=fontTitle, fill=0)
            draw.text((0, 55), 'Owner: ' + fullName,
                      font=fontSubtitle, fill=0)
            draw.text((0, 75), 'Attendance: ' +
                      attendance, font=fontSubtitle, fill=0)
            draw.text((0, 95), 'Points: ' + termPoints,
                      font=fontSubtitle, fill=0)
            draw.text((0, 115), 'Date: ' + currentDay + ' - Week ' +
                      weekNumber, font=fontSubtitle, fill=0)
            draw.text((0, 135), '      ' + formattedDate,
                      font=fontSubtitle, fill=0)

            # Draw the events on the screen.
            draw.line([(0, 165), (400, 165)])
            drawEventsYLocation = 175
            if (calendarData == ""):
                draw.text(
                    (0, drawEventsYLocation), 'There are no events scheduled for today.', font=fontEvents, fill=0)
            for lesson in splitCalendarData:
                draw.text((0, drawEventsYLocation),
                          lesson, font=fontEvents, fill=0)
                drawEventsYLocation += 20

            # Draw the profile picture on the screen.
            profileImage = get_profile_img(headers, arborURL)
            width, height = profileImage.size
            profileImage = profileImage
            virtualImage.paste(profileImage, (270, 0))

            # Display the final image to the screen.
            self.flush_buffer(epd.getbuffer(virtualImage))
            # Send the screen to SLEEP
            self.Send_cmd('S')
        except ConnectionResetError:
            print("The connection to the device was lost.")
        except KeyboardInterrupt:
            print("Keyboard interrupt received, exiting...")
            self.close()


if __name__ == "__main__":
    ip = tcp_sver.get_host_ip()
    print("Server ready for connections!")
    server = socketserver.ThreadingTCPServer((ip, 6868, ), TimetableServer)
    try:
        server.serve_forever()
    except KeyboardInterrupt:
        print("Keyboard interrupt received, exiting...")
        server.shutdown()
