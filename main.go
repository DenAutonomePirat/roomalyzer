package main

import (
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"flag"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/attilabuti/go-snefru"
	"github.com/dnlo/struct2csv"
	"gopkg.in/yaml.v2"
)

func main() {
	var filename string

	flag.StringVar(&filename, "o", "output.csv", "Name of output file")

	flag.Parse()

	//Load config from file
	f, err := os.Open("config.yml")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	var cfg Config
	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(&cfg)
	if err != nil {
		log.Fatal(err)
	}
	f.Close()
	f, err = os.Create(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	u := url.URL{
		Scheme: "https",
		Path:   "app.roomalyzer.com/api/index.php",
	}

	q := u.Query()
	q.Add("lane", "sensor_data")
	q.Add("sensor", cfg.Sensor)
	q.Add("time", strconv.FormatInt(time.Now().Unix(), 10))
	q.Add("lane", "sensor_data")
	q.Add("hours", "48")

	q = addChecksum(q, cfg.Token)
	u.RawQuery = q.Encode()

	response, err := http.Get(u.String())
	if err != nil {
		log.Fatal(err)
	}

	responseData, err := io.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}
	var sensor_data1 Sensor_Data
	err = json.Unmarshal(responseData, &sensor_data1)
	if err != nil {
		log.Fatal(err)
	}

	enc := struct2csv.New()
	rows, err := enc.Marshal(sensor_data1.Data)
	if err != nil {
		log.Fatal(err)
	}
	w := csv.NewWriter(f)

	for _, row := range rows {
		if err := w.Write(row); err != nil {
			log.Fatalln("error writing record to csv:", err)
		}
	}

	// Write any buffered data to the underlying writer
	w.Flush()

	if err := w.Error(); err != nil {
		log.Fatal(err)
	}
}

func addChecksum(q url.Values, token string) url.Values {
	switch lane := q.Get("lane"); lane {
	case "sensor_list":
		/*
			PHP:
			$str = $time.".".$account.".".$token;
			$checksum = hash("snefru256",$str);

		*/
		s := q.Get("time") + "." + q.Get("account") + "." + token

		h := snefru.NewSnefru256(8)
		h.Write([]byte(s))
		q.Add("checksum", hex.EncodeToString(h.Sum(nil)))
		return q
	case "sensor_data":
		/*
			PHP:
			$str = $time.".".$sensor.".".$token;
			$checksum = hash("snefru256",$str);
		*/
		s := q.Get("time") + "." + q.Get("sensor") + "." + token

		h := snefru.NewSnefru256(8)
		h.Write([]byte(s))
		q.Add("checksum", hex.EncodeToString(h.Sum(nil)))
		return q
	default:
		return nil

	}

}

type Sensor_Data struct {
	Status string `json:"status"`
	Data   []struct {
		ID          string `json:"id"`
		Datetime    string `json:"datetime"`
		Sensor      string `json:"sensor"`
		Temperature string `json:"temperature"`
		Humidity    string `json:"humidity"`
		Co2         string `json:"co2"`
		Voc         string `json:"voc"`
		Sound       string `json:"sound"`
		SoundLow    string `json:"sound_low"`
		SoundHigh   string `json:"sound_high"`
		LightLevel  string `json:"light_level"`
		LightColour string `json:"light_colour"`
		Occupancy   string `json:"occupancy"`
		Rssi        string `json:"rssi"`
		Voltage     string `json:"voltage"`
	} `json:"data"`
}

type Config struct {
	Token  string `yaml:"token"`
	Sensor string `yaml:"sensor"`
}
