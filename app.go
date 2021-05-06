package main

import (
	"log"
	"math/rand"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	. "github.com/k4s/phantomgo"
	"github.com/meinside/telegraph-go"
	tb "gopkg.in/tucnak/telebot.v2"
)

// Set Bot's token
var token string

// Set ChatID int
var chatID int

var brower Phantomer
var mangaList map[int]mangaTitle
var chaptersList map[int]string

var baseURL = "https://readmanga.live"
var listURL = "https://readmanga.live/list?sortType=rate&offset="

// Xpath variables
var lastPagePath = ".pagination a.step"
var mangaOnPagePath = ".tiles .tile .desc h3 a"
var listOfChaptersPath = ".table.table-hover tbody a"
var buyButtonPath = "div.flex-row div.subject-meta > a"
var noChapterForReadPath = "div.subject-actions > a"
var mangaPagePicPath = "#fotocontext img"
var countPagesElemPath = ".top-block .pages-count"

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

// Base function that download random chapter of a random manga
func downloadRandomMangaChapter(b *tb.Bot, r tb.Recipient) {
	log.Printf("> Selecting manga ...")
	chaptersList = make(map[int]string)
	mangaNumber := randomizer(len(mangaList))
	selectedManga := mangaList[mangaNumber].Link
	for ChapterListChecker(baseURL+mangaList[mangaNumber].Link) != true {
		chaptersList = make(map[int]string)
		mangaNumber = randomizer(len(mangaList))
		selectedManga = mangaList[mangaNumber].Link
	}
	GetChaptersList(baseURL + selectedManga)
	chapterNumber := randomizer(len(chaptersList))
	selectedChapter := chaptersList[chapterNumber]
	c, f, err := GetCountOfPages(baseURL, selectedChapter)
	for err != nil || c <= 15 {
		if err == nil {
			log.Printf("> Selected manga with error: %s", err)
		}
		log.Printf("> Selected manga pages count: %s", strconv.Itoa(c))
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

// GetMangaList - save all titles and links of a manga list
func GetMangaList() {
	lastPage := CheckLastPageOfMangaList()
	pageNum := randomizer(lastPage)
	mangaList = make(map[int]mangaTitle)
	log.Println("> Parsing manga list started!")
	MangaListParser(0 + (pageNum-1)*70)
	log.Println("> Parsing manga list finished!")
}

// CheckLastPageOfMangaList - Check last page of a manga list
func CheckLastPageOfMangaList() (index int) {
	doc := getPageWithoutJS(listURL + strconv.Itoa(0))
	index, err := strconv.Atoi(doc.Find(lastPagePath).Last().Text())
	checkError(err)
	return
}

// MangaListParser - Parse manga titles list
func MangaListParser(offset int) {
	doc := getPageWithoutJS(listURL + strconv.Itoa(offset))
	doc.Find(mangaOnPagePath).Each(func(i int, s *goquery.Selection) {
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
	t = doc.Find(buyButtonPath).Text()
	if t == "Купить том " {
		isChapterList = false
	} else {
		t = doc.Find(noChapterForReadPath).Text()
		if t != "Читать мангу с первой главы" {
			isChapterList = false
		} else {
			isChapterList = true
		}
	}
	return isChapterList
}

// GetChaptersList - Parse and get list of chapters
func GetChaptersList(URL string) {
	doc := getPageWithoutJS(URL)
	doc.Find(listOfChaptersPath).Each(func(i int, s *goquery.Selection) {
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

// MangaPageParser - Parse and return html page
func MangaPageParser(URL string, selectedChapter string, count int, pre string) string {
	var html string
	log.Println("> Starting concatinate html string")
	resp := getPageWithJS(URL + selectedChapter)
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	checkError(err)
	mangaPageURL, ok := doc.Find(mangaPagePicPath).Attr("src")
	if ok {
		segments := strings.Split(mangaPageURL, "?")
		path := segments[0]
		var re = regexp.MustCompile(`\/\/.*?\.`)
		path = re.ReplaceAllString(path, `//t7.`)
		html = html + `<img src=` + path + `>`
	}
	if pre == "?mtr=1" {
		pre = pre + "#page="
	}
	for i := 0; i < count; i++ {
		resp := getPageWithJS(URL + selectedChapter + pre + strconv.Itoa(i+1))
		doc, err := goquery.NewDocumentFromReader(resp.Body)
		checkError(err)
		mangaPageURL, ok := doc.Find(mangaPagePicPath).Attr("src")
		if ok {
			segments := strings.Split(mangaPageURL, "?")
			path := segments[0]
			var re = regexp.MustCompile(`\/\/.*?\.`)
			path = re.ReplaceAllString(path, `//t7.`)
			html = html + `<img src=` + path + `>`
		}
	}
	log.Println("> Concatinate html string finished!")
	return html
}

// GetCountOfPages - Check count of pages
func GetCountOfPages(URL string, selectedChapter string) (c int, f string, err error) {
	f = "#page="
	resp := getPageWithJS(URL + selectedChapter + f + strconv.Itoa(1))
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	checkError(err)
	c, err = strconv.Atoi(doc.Find(countPagesElemPath).Text())
	if err != nil {
		f = "?mtr=1"
		resp := getPageWithJS(URL + selectedChapter + f)
		doc, err := goquery.NewDocumentFromReader(resp.Body)
		checkError(err)
		log.Println("> Mint manga new URL: " + URL + selectedChapter + "?mtr=" + strconv.Itoa(1))
		c, err = strconv.Atoi(doc.Find(countPagesElemPath).Text())
	}
	return
}

func checkError(err error) {
	if err != nil {
		log.Panic(err)
		panic(err)
	}
}

func getPageWithJS(URL string) *http.Response {
	p := &Param{
		Method:       "GET",
		Url:          URL,
		Header:       http.Header{"Cookie": []string{"your cookie"}},
		UsePhantomJS: true,
	}
	resp, err := brower.Download(p)
	checkError(err)
	return resp
}

func getPageWithoutJS(URL string) *goquery.Document {
	res, err := http.Get(URL)
	checkError(err)
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("* Status code error: %d %s", res.StatusCode, res.Status)
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
	bot, err := tb.NewBot(tb.Settings{
		Token:  token,
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	})
	if err != nil {
		return
	}
	r := tb.Recipient(tb.ChatID(chatID))
	downloadRandomMangaChapter(bot, r)
}
