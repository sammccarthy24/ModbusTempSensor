package main

/*
 * Description:
 * 	This program reads the temperature from  the R414A01 modbus 
 * 	temperature sensor and displays it in the terminal
 * 
*/

import (
	"fmt"
	"bufio"
	"github.com/goburrow/modbus"
	"log"
	"time"
	"os"
	"strconv"
	"gopkg.in/gomail.v2"
)
                
const (
	SAMPLING_PERIOD = 60 // //Seconds
	SERIAL_PORT = "/dev/ttyUSB0"
	FILE_PATH = "/home/pi/Data/Temperature_Data/"
	EMAIL_ADDRESS_FROM = "sammccarthy24.develop@gmail.com"
	EMAIL_ADDRESS_TO = "sammccarthy24@gmail.com"
) 

var (
	EMAIL_PASSWORD_FROM = ""
)


func main() {
	fmt.Println("Modbus temperature sensor data logger")
	
	reader := bufio.NewReader(os.Stdin)
    fmt.Print("Enter password for " + EMAIL_ADDRESS_FROM + ": ")
    EMAIL_PASSWORD_FROM, _ := reader.ReadString('\n')

	// Modbus RTU/ASCII
	handler := modbus.NewRTUClientHandler(SERIAL_PORT)
	handler.BaudRate = 9600
	handler.DataBits = 8
	handler.Parity = "N"
	handler.StopBits = 1
	handler.SlaveId = 1
	handler.Timeout = 5 * time.Second

	err := handler.Connect()
	if err != nil {
			log.Fatal(err)
	}
	defer handler.Close()

	client := modbus.NewClient(handler)

	for ;; {

		// Capture start time
		start_time := []byte(time.Now().Format(time.RFC3339Nano))
		start_time = start_time[0:(len(start_time) - 16)]
		
		// Capture current date
		current_date, _ := strconv.Atoi(string(start_time[8:10]))
		fmt.Println(string(start_time))
		previous_date := current_date

		// Open the file to store the data
		file, err := os.OpenFile(FILE_PATH + string(start_time[:10]) + ".csv", os.O_WRONLY | os.O_CREATE | os.O_APPEND, 0644)
		if err != nil {
				panic(err)
		}
		defer file.Close()

		// Initialise maximum and minimum temperatures
		results, err := client.ReadHoldingRegisters(0, 1)
		if err != nil {
				log.Fatal(err)
		}
		max_temp := float64(int(results[0])*256 + int(results[1]))/10.0
		min_temp := max_temp
	
		for current_date == previous_date {

			// Set previous date to current date for comparison
			previous_date = current_date

			results, err = client.ReadHoldingRegisters(0, 1)
			if err != nil {
					SensorNotFoundError(start_time)
			}

			temp := float64(int(results[0])*256 + int(results[1]))/10.0

			if temp > max_temp {
					max_temp = temp
			} else if temp < min_temp {
					min_temp = temp
			}

			fmt.Print("The temperature is ")
			fmt.Print(temp)
			fmt.Println(" degrees celsius")

			// Capture a time stamp and then re-format it for the .CSV file
			stamp := []byte(time.Now().Format(time.RFC3339Nano))
			for i, c := range stamp {
					if c == 'T' {
							stamp[i] = ','
					}
			}
			stamp = stamp[0:(len(stamp) - 11)]
			stamp = append(stamp, ',')
			
			// Update current date
			current_date, _ = strconv.Atoi(string(stamp[8:10]))

			// Write the time stamp to the .CSV file
			_, err = file.Write(stamp)
			if err != nil {
					panic(err)
			}

			// Write the serial data to the .CSV file
			_, err = file.Write([]byte(strconv.FormatFloat(temp, 'f', 2, 32)))
			if err != nil {
					panic(err)
			}

			// Write the serial data to the .CSV file
			_, err = file.Write([]byte("\n"))
			if err != nil {
					panic(err)
			}

			// Delay x seconds
			time.Sleep(SAMPLING_PERIOD * time.Second)
		}
	
	        // Capture end time
	        end_time := []byte(time.Now().Format(time.RFC3339Nano))
	        end_time = end_time[0:(len(end_time) - 16)]
	
			// Construct email body
	        email_body := "Temperature data recorded from " + string(start_time[:len(start_time)]) + " until " + string(end_time[:len(end_time)]) + ".<br>Maximum temperature was " + strconv.FormatFloat(max_temp, 'f', 1, 32) + " degrees C.<br>Minimum temperature was " + strconv.FormatFloat(min_temp, 'f', 1, 32) + " degrees C."
	
	        m := gomail.NewMessage()
	        m.SetHeader("From", EMAIL_ADDRESS_FROM)
	        m.SetHeader("To", EMAIL_ADDRESS_TO)
	        m.SetHeader("Subject", "Temperature sensor log for " + string(start_time[:len(start_time)-8]))
	        m.SetBody("text/html", email_body)
	        m.Attach(FILE_PATH + string(start_time[:(len(start_time)-6)]) + ".csv")
	
	        d := gomail.NewDialer("smtp.gmail.com", 587, EMAIL_ADDRESS_FROM, EMAIL_PASSWORD_FROM)
	
	
	        // Send the email
	        if err := d.DialAndSend(m); err != nil {
	            panic(err)
	        }
	}
}

func SensorNotFoundError(start_time []byte) {
	
	// Construct error email body
	email_body := "Sensor was not found. Data up until now is attached"

	m := gomail.NewMessage()
	m.SetHeader("From", EMAIL_ADDRESS_FROM)
	m.SetHeader("To", EMAIL_ADDRESS_TO)
	m.SetHeader("Subject", "Temperature sensor error")
	m.SetBody("text/html", email_body)
	m.Attach(FILE_PATH + string(start_time[:(len(start_time)-6)]) + ".csv")

	d := gomail.NewDialer("smtp.gmail.com", 587, EMAIL_ADDRESS_FROM, EMAIL_PASSWORD_FROM)


	// Send the email
	if err := d.DialAndSend(m); err != nil {
		panic(err)
	}
	
	log.Fatal()
}
