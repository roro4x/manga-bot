package main

import (
	"log"
	"math/rand"
	"net/http"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/jasonlvhit/gocron"
	"github.com/meinside/telegraph-go"
	. "github.com/roro4x/phantomgo"
	tb "gopkg.in/tucnak/telebot.v2"
)

var token = "TOKEN"
var chatID int // = int

var brower Phantomer

var baseURL = "https://readmanga.live"
var listURL = "https://readmanga.live/list?sortType=rate&offset="

var mangaList map[int]mangaTitle
var chaptersList map[int]string

type mangaTitle struct {
	Title string `json:"title"`
	Link  string `json:"link"`
}

// Create Telegraph page with manga chapter
func createTelegraphPage(html string, title string) string {
	const verbose = true
	telegraph.Verbose = verbose
	var savedAccessToken string
	var pageURL string

	if client, err := telegraph.Create("telegraph-mangakaka", "Mangakaka", ""); err == nil {
		savedAccessToken = client.AccessToken

		// CreatePage
		if page, err := client.CreatePageWithHTML(title, "Mangakaka", "", html, true); err == nil {
			pageURL = page.URL
			log.Printf("> CreatePage result: %#+v", page)
			log.Printf("> Created page url: %s", page.URL)
		} else {
			log.Printf("* CreatePage error: %s", err)
		}
	} else {
		log.Printf("* Create error: %s", err)
	}
	if savedAccessToken == "" {
		log.Printf("* Couldn't save access token, exiting...")
		return pageURL
	}
	return pageURL
}

// Send Manga Chapter to Telegram channel
func sendMangaChapter(b *tb.Bot, r tb.Recipient, t string) {
	b.Send(r, t)
}

// Base function that download random chapter of a random manga.
func downloadRandomMangaChapter(b *tb.Bot, r tb.Recipient) {
	chaptersList = make(map[int]string)
	mangaNumber := randomizer(len(mangaList))
	selectedManga := mangaList[mangaNumber].Link
	for ChapterListChecker(baseURL+selectedManga) != true {
		chaptersList = make(map[int]string)
		mangaNumber = randomizer(len(mangaList))
		selectedManga = mangaList[mangaNumber].Link
	}
	GetChaptersList(baseURL + selectedManga)
	chapterNumber := randomizer(len(chaptersList))
	selectedChapter := chaptersList[chapterNumber]
	c, f, err := GetCountOfPages(baseURL, selectedChapter)
	for err != nil || c <= 15 {
		chaptersList = make(map[int]string)
		mangaNumber = randomizer(len(mangaList))
		selectedManga = mangaList[mangaNumber].Link
		for ChapterListChecker(baseURL+selectedManga) != true {
			chaptersList = make(map[int]string)
			mangaNumber = randomizer(len(mangaList))
			selectedManga = mangaList[mangaNumber].Link
		}
		GetChaptersList(baseURL + selectedManga)
		chapterNumber = randomizer(len(chaptersList))
		selectedChapter = chaptersList[chapterNumber]
		c, f, err = GetCountOfPages(baseURL, selectedChapter)
	}
	log.Println("> Selected manga: " + baseURL + selectedChapter + "#page=")
	html := MangaPageParser(baseURL, selectedChapter, c, f)
	pageURL := createTelegraphPage(`<a href="`+baseURL+mangaList[mangaNumber].Link+`">`+mangaList[mangaNumber].Title+`</a>`+html+`<a href="https://readmanga.live">Читать мангу Online</a>`, mangaList[mangaNumber].Title)
	sendMangaChapter(b, r, pageURL)
}

// GetMangaList - save all titles and links of a manga list.
func GetMangaList() {
	lastPage := CheckLastPageOfMangaList()
	mangaList = make(map[int]mangaTitle)
	log.Println("> Parsing manga list started!")
	offset := 0
	for i := 1; i <= lastPage; i++ {
		if i%70 == 0 {
			time.Sleep(2 * time.Minute)
		}
		MangaListParser(offset)
		offset = offset + 70
	}
	log.Println("> Parsing manga list finished!")
}

// CheckLastPageOfMangaList - Check last page of a manga list.
func CheckLastPageOfMangaList() (index int) {
	doc := getPageWithoutJS(listURL + strconv.Itoa(0))
	index, err := strconv.Atoi(doc.Find(".pagination a.step").Last().Text())
	checkError(err)
	return
}

// MangaListParser - Parse manga titles list.
func MangaListParser(offset int) {
	doc := getPageWithoutJS(listURL + strconv.Itoa(offset))
	doc.Find(".tiles.row .tile.col-sm-6 .desc h3 a").Each(func(i int, s *goquery.Selection) {
		title := s.Text()
		link, ok := s.Attr("href")
		if ok {
			mangaList[len(mangaList)] = mangaTitle{
				Title: title,
				Link:  link,
			}
		}
	})
}

// ChapterListChecker - Check is need to buy manga or manga have chapters
func ChapterListChecker(URL string) bool {
	var isChapterList bool
	var t string
	doc := getPageWithoutJS(URL)
	t = doc.Find("div.flex-row div.subject-meta > a").Text()
	if t == "Купить том " {
		isChapterList = false
	} else {
		t = doc.Find("div.subject-actions > a").Text()
		if t != "Читать мангу с первой главы" {
			isChapterList = false
		} else {
			isChapterList = true
		}
	}
	return isChapterList
}

// GetChaptersList - Parse and get list of chapters.
func GetChaptersList(URL string) {
	doc := getPageWithoutJS(URL)
	doc.Find(".table.table-hover tbody a").Each(func(i int, s *goquery.Selection) {
		link, ok := s.Attr("href")
		if ok {
			chaptersList[i] = link
		} else {
			mangaNumber := randomizer(len(mangaList))
			selectedManga := mangaList[mangaNumber].Link
			GetChaptersList(baseURL + selectedManga)
		}
	})
}

// MangaPageParser - Parse and return html page.
func MangaPageParser(URL string, selectedChapter string, count int, pre string) string {
	var html string
	log.Println("> Starting concatinate html string")
	if pre == "#page=" {
		resp, cmd := getPageWithJS(URL + selectedChapter)
		doc, err := goquery.NewDocumentFromReader(resp.Body)
		checkError(err)
		mangaPageURL, ok := doc.Find("#fotocontext img").Attr("src")
		if ok {
			segments := strings.Split(mangaPageURL, "?")
			path := segments[0]
			var re = regexp.MustCompile(`\/\/.*?\.`)
			path = re.ReplaceAllString(path, `//t7.`)
			html = html + `<img src=` + path + `>`
		}
		brower.Close(cmd)
	}
	if pre == "?mtr=1" {
		resp, cmd := getPageWithJS(URL + selectedChapter)
		doc, err := goquery.NewDocumentFromReader(resp.Body)
		checkError(err)
		mangaPageURL, ok := doc.Find("#fotocontext img").Attr("src")
		if ok {
			segments := strings.Split(mangaPageURL, "?")
			path := segments[0]
			var re = regexp.MustCompile(`\/\/.*?\.`)
			path = re.ReplaceAllString(path, `//t7.`)
			html = html + `<img src=` + path + `>`
		}
		pre = pre + "#page="
		brower.Close(cmd)
	}
	for i := 0; i < count; i++ {
		resp, cmd := getPageWithJS(URL + selectedChapter + pre + strconv.Itoa(i+1))
		doc, err := goquery.NewDocumentFromReader(resp.Body)
		checkError(err)
		mangaPageURL, ok := doc.Find("#fotocontext img").Attr("src")
		if ok {
			segments := strings.Split(mangaPageURL, "?")
			path := segments[0]
			var re = regexp.MustCompile(`\/\/.*?\.`)
			path = re.ReplaceAllString(path, `//t7.`)
			html = html + `<img src=` + path + `>`
		}
		brower.Close(cmd)

	}
	log.Println("> Concatinate html string finished!")
	return html
}

// GetCountOfPages - Check count of pages.
func GetCountOfPages(URL string, selectedChapter string) (c int, f string, err error) {
	f = "#page="
	resp, cmd := getPageWithJS(URL + selectedChapter + f + strconv.Itoa(1))
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	checkError(err)
	c, err = strconv.Atoi(doc.Find(".top-block .pages-count").Text())
	if err != nil {
		brower.Close(cmd)
		f = "?mtr=1"
		resp, cmd = getPageWithJS(URL + selectedChapter + f)
		doc, err := goquery.NewDocumentFromReader(resp.Body)
		checkError(err)
		log.Println("> Mint manga new URL: " + URL + selectedChapter + "?mtr=" + strconv.Itoa(1))
		c, err = strconv.Atoi(doc.Find(".top-block .pages-count").Text())
	}
	brower.Close(cmd)
	return
}

func checkError(err error) {
	if err != nil {
		log.Panic(err)
		panic(err)
	}
}

func getPageWithJS(URL string) (*http.Response, *exec.Cmd) {
	p := &Param{
		Method:       "GET",
		Url:          URL,
		Header:       http.Header{"Cookie": []string{"your cookie"}},
		UsePhantomJS: true,
	}
	resp, err, cmd := brower.Download(p)
	checkError(err)
	return resp, cmd
}

func getPageWithoutJS(URL string) *goquery.Document {
	res, err := http.Get(URL)
	checkError(err)
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}
	doc, err := goquery.NewDocumentFromReader(res.Body)
	checkError(err)
	return doc
}

func randomizer(i int) int {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(i)
}

func main() {
	brower = NewPhantom()
	GetMangaList()
	b, err := tb.NewBot(tb.Settings{
		Token:  token,
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	})
	if err != nil {
		return
	}
	r := tb.Recipient(tb.ChatID(chatID))
	gocron.Every(30).Minutes().From(gocron.NextTick()).Do(downloadRandomMangaChapter, b, r)
	<-gocron.Start()
	b.Start()
}
