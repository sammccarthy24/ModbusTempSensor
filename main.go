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
	SAMPLING_PERIOD = 60 // Seconds
	SERIAL_PORT = "/dev/ttyUSB0"
	FILE_PATH = "/home/pi/Data/Temperature_Data/"
	EMAIL_ADDRESS_FROM = "sammccarthy24.develop@gmail.com"
	EMAIL_ADDRESS_TO = "sammccarthy24@gmail.com"
) 

var (
	EMAIL_PASSWORD_FROM = ""
	FILE_NAME = ""
)


func main() {
	fmt.Println("Modbus temperature sensor data logger")
	
	reader := bufio.NewReader(os.Stdin)
    fmt.Print("Enter password for " + EMAIL_ADDRESS_FROM + ": ")
    EMAIL_PASSWORD_FROM, _ := reader.ReadString('\n')

	// Modbus RTU/ASCII: Connect to sensor 1
	handler1 := modbus.NewRTUClientHandler(SERIAL_PORT)
	handler1.BaudRate = 9600
	handler1.DataBits = 8
	handler1.Parity = "N"
	handler1.StopBits = 1
	handler1.SlaveId = 1
	handler1.Timeout = 5 * time.Second

	err := handler1.Connect()
	if err != nil {
			log.Fatal(err)
	}
	defer handler1.Close()

	client1 := modbus.NewClient(handler1)
	
	// Modbus RTU/ASCII: Connect to sensor 2
	handler2 := modbus.NewRTUClientHandler(SERIAL_PORT)
	handler2.BaudRate = 9600
	handler2.DataBits = 8
	handler2.Parity = "N"
	handler2.StopBits = 1
	handler2.SlaveId = 2
	handler2.Timeout = 5 * time.Second

	err = handler2.Connect()
	if err != nil {
			log.Fatal(err)
	}
	defer handler2.Close()

	client2 := modbus.NewClient(handler2)
	
	// Modbus RTU/ASCII: Connect to sensor 3
	handler3 := modbus.NewRTUClientHandler(SERIAL_PORT)
	handler3.BaudRate = 9600
	handler3.DataBits = 8
	handler3.Parity = "N"
	handler3.StopBits = 1
	handler3.SlaveId = 3
	handler3.Timeout = 5 * time.Second

	err = handler3.Connect()
	if err != nil {
			log.Fatal(err)
	}
	defer handler3.Close()

	client3 := modbus.NewClient(handler3)
	

	for ;; {

		// Capture start time
		start_time := []byte(time.Now().Format(time.RFC3339Nano))
		start_time = start_time[0:(len(start_time) - 16)]
		
		// Capture current date
		current_date, _ := strconv.Atoi(string(start_time[8:10]))
		fmt.Println(string(start_time))
		previous_date := current_date

		// Open the file to store the data
		FILE_NAME = string(start_time[:10]) + ".csv"
		file, err := os.OpenFile(FILE_PATH + FILE_NAME, os.O_WRONLY | os.O_CREATE | os.O_APPEND, 0644)
		if err != nil {
				panic(err)
		}
		defer file.Close()

		// Initialise maximum and minimum temperatures
		results, err := client1.ReadHoldingRegisters(0, 1)
		if err != nil {
				log.Fatal(err)
		}
		max_temp := float64(int(results[0])*256 + int(results[1]))/10.0
		min_temp := max_temp
	
		for current_date == previous_date {

			// Set previous date to current date for comparison
			previous_date = current_date

			// Read sensor 1
			results, err = client1.ReadHoldingRegisters(0, 1)
			if err != nil {
					SensorNotFoundError(start_time)
			}
			temp := float64(int(results[0])*256 + int(results[1]))/10.0
			
			// Read sensor 2
			results, err = client2.ReadHoldingRegisters(0, 1)
			if err != nil {
					SensorNotFoundError(start_time)
			}
			temp2 := float64(int(results[0])*256 + int(results[1]))/10.0
			
			// Read sensor 3
			results, err = client3.ReadHoldingRegisters(0, 1)
			if err != nil {
					SensorNotFoundError(start_time)
			}
			temp3 := float64(int(results[0])*256 + int(results[1]))/10.0

			if temp > max_temp {
					max_temp = temp
			} else if temp < min_temp {
					min_temp = temp
			}

			fmt.Print("The temperature on sensor 1 is ")
			fmt.Print(temp)
			fmt.Println(" degrees celsius")
			
			fmt.Print("The temperature on sensor 2 is ")
			fmt.Print(temp2)
			fmt.Println(" degrees celsius")
			
			fmt.Print("The temperature on sensor 3 is ")
			fmt.Print(temp3)
			fmt.Println(" degrees celsius\n")

			// Capture a time stamp and then re-format it for the .CSV file
			stamp := []byte(time.Now().Format(time.RFC3339Nano))
			for i, c := range stamp {
					if c == 'T' {
							stamp[i] = ','
					}
			}
			stamp = stamp[0:19]
			stamp = append(stamp, ',')
			
			// Update current date
			current_date, _ = strconv.Atoi(string(stamp[8:10]))
			
			// Build entry string
			entry := string(stamp) + strconv.FormatFloat(temp, 'f', 2, 32) + "," + strconv.FormatFloat(temp2, 'f', 2, 32) + "," + strconv.FormatFloat(temp3, 'f', 2, 32) + "\n" 

			// Write the time stamp to the .CSV file
			_, err = file.Write([]byte(entry))
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
	        m.SetHeader("Subject", "Temperature sensor log for " + string(start_time[:len(start_time)-9]))
	        m.SetBody("text/html", email_body)
	        m.Attach(FILE_PATH + FILE_NAME)
	
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
	m.Attach(FILE_PATH + FILE_NAME)

	d := gomail.NewDialer("smtp.gmail.com", 587, EMAIL_ADDRESS_FROM, EMAIL_PASSWORD_FROM)


	// Send the email
	if err := d.DialAndSend(m); err != nil {
		panic(err)
	}
	
	log.Fatal()
}
