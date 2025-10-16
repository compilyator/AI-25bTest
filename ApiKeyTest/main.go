package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"

	"github.com/logrusorgru/aurora/v4"
)

// PexelsResponse описує відповідь API Pexels
type PexelsResponse struct {
	TotalResults int     `json:"total_results"`
	Page         int     `json:"page"`
	PerPage      int     `json:"per_page"`
	Photos       []Photo `json:"photos"`
	NextPage     string  `json:"next_page"`
}

// Photo описує фото у відповіді
type Photo struct {
	ID              int      `json:"id"`
	Width           int      `json:"width"`
	Height          int      `json:"height"`
	URL             string   `json:"url"`
	Photographer    string   `json:"photographer"`
	PhotographerURL string   `json:"photographer_url"`
	PhotographerID  int      `json:"photographer_id"`
	AvgColor        string   `json:"avg_color"`
	Src             PhotoSrc `json:"src"`
	Liked           bool     `json:"liked"`
	Alt             string   `json:"alt"`
}

// PhotoSrc описує різні розміри фото
type PhotoSrc struct {
	Original  string `json:"original"`
	Large2x   string `json:"large2x"`
	Large     string `json:"large"`
	Medium    string `json:"medium"`
	Small     string `json:"small"`
	Portrait  string `json:"portrait"`
	Landscape string `json:"landscape"`
	Tiny      string `json:"tiny"`
}

func main() {
	// Завантажуємо .env
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Не вдалося завантажити .env файл:", err)
		return
	}
	apiKey := os.Getenv("PEXELS_API_KEY")
	if apiKey == "" {
		fmt.Println("PEXELS_API_KEY не знайдено у .env!")
		return
	}
	apiServer := "https://api.pexels.com/v1"

	var query string
	fmt.Print("Введіть параметр пошуку фото: ")
	fmt.Scanln(&query)
	if query == "" {
		fmt.Println("Параметр пошуку не може бути порожнім!")
		return
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/search?query=%s", apiServer, query), nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}
	req.Header.Set("Authorization", apiKey)

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error making request:", err)
		return
	}
	defer resp.Body.Close()

	// Витягуємо rate limit статистику з заголовків
	limit := resp.Header.Get("X-Ratelimit-Limit")
	remaining := resp.Header.Get("X-Ratelimit-Remaining")
	reset := resp.Header.Get("X-Ratelimit-Reset")

	if resp.StatusCode == 429 {
		fmt.Println("Too many requests. Please try again later.")
		return
	} else if resp.StatusCode != 200 {
		fmt.Println("Error: received status code", resp.StatusCode)
		return
	}

	var pexelsResponse PexelsResponse
	if err := json.NewDecoder(resp.Body).Decode(&pexelsResponse); err != nil {
		fmt.Println("Error decoding response:", err)
		return
	}

	for index, photo := range pexelsResponse.Photos {
		photographer := aurora.BrightBlue(photo.Photographer).Bold()
		description := aurora.BrightMagenta(photo.Alt).Italic()
		// Гіперпосилання для консолі Windows (PowerShell):
		link := aurora.BrightCyan("URL").Hyperlink(photo.URL)
		largeUrl := aurora.Green("Large").Hyperlink(photo.Src.Large)
		fmt.Printf("%d %s: %s %s %s\n", index+1, photographer, description, link, largeUrl)
	}

	// Вивід статистики по rate limit
	fmt.Println(aurora.Yellow("Статистика по обмеженням API:").Bold().String())
	fmt.Printf("Всього запитів на місяць: %s\n", limit)
	fmt.Printf("Залишилось запитів: %s\n", remaining)
	// Перетворення UNIX timestamp у дату
	var resetDate string
	if reset != "" {
		if ts, err := parseUnixTimestamp(reset); err == nil {
			resetDate = ts
		} else {
			resetDate = reset + " (невірний формат)"
		}
	} else {
		resetDate = "невідомо"
	}
	fmt.Printf("Оновлення ліміту: %s\n", resetDate)

}

// parseUnixTimestamp перетворює UNIX timestamp у людинозчитувану дату
func parseUnixTimestamp(ts string) (string, error) {
	var unix int64
	_, err := fmt.Sscanf(ts, "%d", &unix)
	if err != nil {
		return "", err
	}
	t := unixToTime(unix)
	return t.Format("02.01.2006 15:04:05 MST"), nil
}

func unixToTime(unix int64) (t time.Time) {
	return time.Unix(unix, 0).Local()
}
