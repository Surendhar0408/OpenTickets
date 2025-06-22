package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	_ "io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	_ "net/http/cookiejar"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gomail"
)

var temp interface{}

type Urls struct {
	login     string
	today     string
	wip       string
	awaiting  string
	onhold    string
	awaiting3 string
}

type Detail struct {
	Id         interface{} `json:"id"`
	Status     interface{} `json:"status"`
	Title      interface{} `json:"title"`
	Company    interface{} `json:"company"`
	ReqTime    interface{} `json:"req_time"`
	ReqTime1   interface{} `json:"req_time1"`
	Assignedto interface{} `json:"responsibility"`
}

var wholeTicket []Detail

var write, alltickets [][]string
var w []string
var company int
var flepath string

func main() {

	U := Urls{
		login:     "https://crayonte.sysaidit.com/api/v1/login",
		today:     "https://crayonte.sysaidit.com/api/v1/sr/?fields=title,insert_time,company,status&limit=30&sort=id&dir=desc",
		wip:       "https://crayonte.sysaidit.com/api/v1/sr/?fields=title,insert_time,company,status&sort=id&dir=desc&status=5",
		awaiting:  "https://crayonte.sysaidit.com/api/v1/sr/?fields=title,insert_time,company,status&sort=id&dir=desc&status=41",
		awaiting3: "https://crayonte.sysaidit.com/api/v1/sr/?fields=title,insert_time,company,status&sort=id&dir=desc&status=42",
		onhold:    "https://crayonte.sysaidit.com/api/v1/sr/?fields=title,insert_time,company,status&sort=id&dir=desc&status=43",
	}

	w = []string{"Company/Status", "New", "WorkInProgress", "Awaiting", "Awaiting3rd", "OnHold", ">14Days", ">7Days", ">7Days Infra"}
	write = append(write, w)

	values := map[string]string{"user_name": "surendhar.balaji@crayonte.com", "password": ""}
	json_data, err := json.Marshal(values)

	if err != nil {
		log.Fatal(err)
	}

	resp, err := http.Post(U.login, "application/json",
		bytes.NewBuffer(json_data))

	if err != nil {
		log.Fatal(err)
	}
	jar, err := cookiejar.New(nil)
	if err != nil {
		log.Fatalf("Got error while creating cookie jar %s", err.Error())
	}

	client := http.Client{
		Jar: jar,
	}

	//-------------------Today's ticket------------------------------------------

	var todayDetail []Detail
	urlObj, _ := url.Parse(U.today)
	client.Jar.SetCookies(urlObj, resp.Cookies())

	srresp, err := http.NewRequest("GET", U.today, nil)
	todayResp, err := client.Do(srresp)
	todayDetail = ResponsetoStruct(todayResp)
	TodaySeperator(todayDetail)

	//------------------Work In Progress-----------------------------------------

	var wipDetail []Detail
	urlObj, _ = url.Parse(U.wip)
	client.Jar.SetCookies(urlObj, resp.Cookies())

	srresp, err = http.NewRequest("GET", U.wip, nil)
	wipResp, err := client.Do(srresp)

	wipDetail = ResponsetoStruct(wipResp)
	WholeTickets(wipDetail)

	//------------------Awaiting Customer--------------------------------------------

	var awaitDetail []Detail
	urlObj, _ = url.Parse(U.awaiting)
	client.Jar.SetCookies(urlObj, resp.Cookies())

	srresp, err = http.NewRequest("GET", U.awaiting, nil)
	awaitResp, err := client.Do(srresp)

	awaitDetail = ResponsetoStruct(awaitResp)

	WholeTickets(awaitDetail)

	//-------------------All Tickets--------------------------------------------------
	var ticket []string
	alltickets = append(alltickets, ticket)
	alltickets = append(alltickets, ticket)
	for _, val := range wholeTicket {
		ticket = append(ticket, time.Now().Format("02-01-2006"), fmt.Sprintf("%s", val.Id), fmt.Sprintf("%s", val.Title), fmt.Sprintf("%s", val.Status), fmt.Sprintf("%s", val.Company))
		alltickets = append(alltickets, ticket)
		ticket = []string{}
	}

	//-------------------Awaiting 3rd Party-------------------------------------------

	var await3Detail []Detail
	urlObj, _ = url.Parse(U.awaiting3)
	client.Jar.SetCookies(urlObj, resp.Cookies())

	srresp, err = http.NewRequest("GET", U.awaiting3, nil)
	await3Resp, err := client.Do(srresp)

	await3Detail = ResponsetoStruct(await3Resp)
	WholeTickets(await3Detail)

	//---------------------On Hold-------------------------------------------------------

	var holdDetail []Detail
	urlObj, _ = url.Parse(U.onhold)
	client.Jar.SetCookies(urlObj, resp.Cookies())

	srresp, err = http.NewRequest("GET", U.onhold, nil)
	holdResp, err := client.Do(srresp)

	holdDetail = ResponsetoStruct(holdResp)
	WholeTickets(holdDetail)

	//----------------------14,7 Days---------------------------------------------------------

	DaysOperation(client, resp)

	//---------------------PivotTable-------------------------------------------------------

	CompanySeperation(wholeTicket)

	//---------------------All Tickets to write --------------------------------------------

	for _, val := range alltickets {
		write = append(write, val)
	}

	//---------------------CSV Write---------------------------------------------------------

	CsvWrite(write)

	//---------------------Mail-------------------------------------------------------------

	mail()
}
func ResponsetoStruct(resp *http.Response) []Detail {

	var d Detail
	var darr []Detail
	body, _ := ioutil.ReadAll(resp.Body)
	_ = json.Unmarshal(body, &temp)
	temp0 := temp.([]interface{})

	for _, val := range temp0 {

		temp1, _ := val.(map[string]interface{})

		d = InfotoDetail(temp1["info"])

		d.Id = temp1["id"]

		darr = append(darr, d)

	}

	return darr
}
func InfotoDetail(Info interface{}) Detail {
	var d Detail
	Infoarr := Info.([]interface{})

	for _, val := range Infoarr {

		info_map := val.(map[string]interface{})

		switch info_map["key"] {

		case "status":
			d.Status = info_map["valueCaption"]
			continue
		case "company":
			d.Company = info_map["valueCaption"]
			continue
		case "title":
			d.Title = info_map["valueCaption"]
			continue
		case "insert_time":

			d.ReqTime = info_map["valueCaption"]
			d.ReqTime1 = info_map["value"]
			continue
		case "responsibility":
			d.Assignedto = info_map["valueCaption"]

		}

	}
	return d
}
func CompanySeperation(details []Detail) {
	var cpfi, cray, jch, khind, melm, rsi, sanden, sun []Detail
	for _, detail := range details {
		switch detail.Company {
		case "Century Pacific Food, Inc.":
			cpfi = append(cpfi, detail)

			continue
		case "Crayonte":
			cray = append(cray, detail)

			continue
		case "Johnson Controls Hitachi Air Conditioning Malaysia Sdn Bhd":
			jch = append(jch, detail)

			continue
		case "Khind Holdings Berhad":
			khind = append(khind, detail)

			continue
		case "Mitsubishi Elevator Malaysia Sdn Bhd":
			melm = append(melm, detail)

			continue
		case "Royal Selangor International Sdn Bhd":
			rsi = append(rsi, detail)

			continue
		case "Sanden Air Conditioning (Malaysia) Sdn Bhd":
			sanden = append(sanden, detail)

			continue
		case "Sunchirin Industries (Malaysia) Berhad":
			sun = append(sun, detail)
			continue

		}
	}
	var wholecomp [][]Detail
	wholecomp = append(wholecomp, cpfi, cray, jch, khind, melm, rsi, sanden, sun)
	for _, val := range wholecomp {

		arrval := PivotTable(val)
		write = append(write, arrval)
	}
	//fmt.Println(write)

}
func CsvWrite(write [][]string) {
	csv1, _, filepath := fileCreation()
	writer := csv.NewWriter(csv1)
	for _, pr := range write {

		err := writer.Write(pr)
		if err != nil {
			fmt.Println(err)
		}

	}
	writer.Flush()
	csv1.Close()
	flepath = filepath
}
func fileCreation() (*os.File, *os.File, string) {

	date := time.Now().Format("02-01-2006")
	date = date + " OpenTickets"
	filepath := fmt.Sprintf("C:\\Users\\lnadmin\\Desktop\\OpenTickets\\Opentickets\\%s.csv", date)

	//C:\\Users\\lnadmin\\Desktop\\OpenTickets\\Opentickets --internal
	//C:\\Surendhar\\AMS --Local

	fileout, _ := os.Create(filepath)

	return fileout, nil, filepath
}
func PivotTable(detail []Detail) []string {
	var arrval, fourteen, seven, seveninfra []string
	var comp string
	var new1, wip, await, await3, onhold int
	company++
	for _, val := range detail {
		comp = fmt.Sprintf("%s", val.Company)
		switch val.Status {
		case "New":
			new1++
			continue
		case "Work in Progress":
			wip++
			continue
		case "Awaiting Customer":
			await++
			continue
		case "Awaiting 3rd Party":
			await3++
			continue
		case "On Hold":
			onhold++
			continue
		case ">14Days":
			fourteen = append(fourteen, fmt.Sprintf("#%s", val.Id))
		case ">7Days":
			seven = append(seven, fmt.Sprintf("#%s", val.Id))
		case ">7Days Infra":
			seveninfra = append(seveninfra, fmt.Sprintf("#%s", val.Id))

		}
	}
	s := strings.Join(fourteen, ",")
	s7 := strings.Join(seven, ",")
	s7i := strings.Join(seveninfra, ",")
	if comp == "" {
		arrval = append(arrval, companyName(company), fmt.Sprintf("%d", new1), fmt.Sprintf("%d", wip), fmt.Sprintf("%d", await), fmt.Sprintf("%d", await3), fmt.Sprintf("%d", onhold), s, s7, s7i)

	} else {
		arrval = append(arrval, comp, fmt.Sprintf("%d", new1), fmt.Sprintf("%d", wip), fmt.Sprintf("%d", await), fmt.Sprintf("%d", await3), fmt.Sprintf("%d", onhold), s, s7, s7i)

	}
	return arrval
}
func TodaySeperator(todayDetail []Detail) []Detail {
	var todayDtl []Detail
	today := time.Now().Format("02-01-2006")

	for _, val := range todayDetail {
		reqdate := fmt.Sprintf("%s", val.ReqTime)
		reqdate = reqdate[0:10]

		if strings.TrimSpace(today) != strings.TrimSpace(reqdate) {
			//fmt.Println(todayDtl)
			return todayDtl
		} else {
			val.Status = "New"
			todayDtl = append(todayDtl, val)
			wholeTicket = append(wholeTicket, val)
		}
	}

	return todayDtl
}
func WholeTickets(tickets []Detail) []Detail {
	for _, val := range tickets {
		wholeTicket = append(wholeTicket, val)
	}
	return wholeTicket
}
func DaysOperation(client http.Client, resp *http.Response) {

	//---------------------------------14--------------------------------------------------------------

	t := time.Now().AddDate(0, 0, -13)
	fourteen := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, t.Nanosecond(), t.Location())

	//fmt.Println(fourteen.UnixNano() / int64(time.Millisecond))
	fourteenurl := fmt.Sprintf("https://crayonte.sysaidit.com/api/v1/sr/?fields=title,insert_time,company,status,responsibility&sort=insert_time&status=5,41,42&insert_time=0,%s",
		fmt.Sprintf("%d", fourteen.UnixNano()/int64(time.Millisecond)))
	urlObj, _ := url.Parse(fourteenurl)
	client.Jar.SetCookies(urlObj, resp.Cookies())

	fourteensrreq, _ := http.NewRequest("GET", fourteenurl, nil)
	fourteenResp, _ := client.Do(fourteensrreq)

	fourteenDetail := ResponsetoStruct(fourteenResp)

	for _, val := range fourteenDetail {
		if val.Assignedto == "Cloud Compute Services" {

			continue
		}
		val.Status = ">14Days"
		wholeTicket = append(wholeTicket, val)
		//fmt.Println("14",val.Id)
	}

	//-------------------------------------7--------------------------------------------------------------

	t7 := time.Now().AddDate(0, 0, -6)
	seven := time.Date(t7.Year(), t7.Month(), t7.Day(), 0, 0, 0, t7.Nanosecond(), t7.Location())
	//fmt.Println(seven.UnixNano() / int64(time.Millisecond))

	sevenurl := fmt.Sprintf("https://crayonte.sysaidit.com/api/v1/sr/?fields=title,insert_time,company,status,responsibility&sort=insert_time&status=5,41,42&insert_time=0,%s", fmt.Sprintf("%d", seven.UnixNano()/int64(time.Millisecond)))
	fmt.Println(sevenurl)
	urlObj, _ = url.Parse(sevenurl)
	client.Jar.SetCookies(urlObj, resp.Cookies())

	sevensrreq, _ := http.NewRequest("GET", sevenurl, nil)
	sevenResp, _ := client.Do(sevensrreq)

	sevenDetail := ResponsetoStruct(sevenResp)

	for i, val := range sevenDetail {

		if val.Assignedto == "Cloud Compute Services" {
			val.Status = ">7Days Infra"
			wholeTicket = append(wholeTicket, val)
			continue
		}
		if i > len(fourteenDetail)-1 {
			val.Status = ">7Days"

			wholeTicket = append(wholeTicket, val)
		}

	}

}
func companyName(i int) string {
	switch i {
	case 1:
		return "Century Pacific Food, Inc."
	case 2:
		return "Crayonte"
	case 3:
		return "Johnson Controls Hitachi Air Conditioning Malaysia Sdn Bhd"
	case 4:
		return "Khind Holdings Berhad"
	case 5:
		return "Mitsubishi Elevator Malaysia Sdn Bhd"
	case 6:
		return "Royal Selangor International Sdn Bhd"
	case 7:
		return "Sanden Air Conditioning (Malaysia) Sdn Bhd"
	case 8:
		return "Sunchirin Industries (Malaysia) Berhad"

	}
	return ""
}
func mail() {
	m := gomail.NewMessage()
	m.SetHeader("From", "surendhar.balaji@crayonte.com")
	m.SetHeader("To", "ams@crayonte.com")
	//m.SetAddressHeader("Cc", "rekha.sanmugam@crayonte.com", "Rekha")
	m.SetHeader("Subject", fmt.Sprintf("OpenTickets Week %d", weeknumber()))
	m.SetBody("text/html", "This is an auto generated mail for <b>Open Tickets<b> file.")
	//m.SetBody("text/html","Just copy the content and paste in Crayonte OpenTickets-2022 sheet.")
	m.Attach(flepath)

	d := gomail.NewPlainDialer("smtp.gmail.com", 587, "surendhar.balaji@crayonte.com", "*******")

	// Send the email to Bob, Cora and Dan.
	if err := d.DialAndSend(m); err != nil {
		panic(err)
	}
}
func weeknumber() int {
	tn := time.Now() // change this to calculate different year

	fmt.Println("Now : ", tn.Format(time.ANSIC))

	// this week is which number in the current calender year
	_, currentWeekNumber := tn.ISOWeek()
	fmt.Println("Current week number : ", currentWeekNumber)
	return currentWeekNumber
}
