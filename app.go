package main

import (
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/jasonlvhit/gocron"
	. "github.com/k4s/phantomgo"
	tb "gopkg.in/tucnak/telebot.v2"
)

var token = "YOUR_TOKEN"
var chatID = "YOUR_CHAT_ID"

var baseURL = "https://readmanga.live"
var listURL = "https://readmanga.live/list?sortType=rate&offset="
var mangaList map[int]mangaTitle
var chaptersList map[int]string

type mangaTitle struct {
	Title string `json:"title"`
	Link  string `json:"link"`
}

// Send Manga Chapter to Telegram channel
func sendMangaChapter(b *tb.Bot, r tb.Recipient, t string) {
	b.Send(r, t)
	files := getListFilesName()
	var a tb.Album
	for i := 0; i < len(files); i++ {
		if len(a) == 9 || i == len(files)-1 {
			a = append(a, &tb.Photo{File: tb.FromDisk("chapter/" + files[i])})
			b.SendAlbum(r, a)
			log.Println("Sent 10 or less chaters")
			a = a[:0]
			time.Sleep(1 * time.Minute)
			continue
		}
		a = append(a, &tb.Photo{File: tb.FromDisk("chapter/" + files[i])})
	}
}

// Base function that download random chapter of a random manga.
func downloadRandomMangaChapter(b *tb.Bot, r tb.Recipient) {
	chaptersList = make(map[int]string)
	mangaNumber := randomizer(len(mangaList))
	selectedManga := mangaList[mangaNumber].Link
	GetChaptersList(baseURL + selectedManga)
	chapterNumber := randomizer(len(chaptersList))
	selectedChapter := chaptersList[chapterNumber]
	log.Println(baseURL + selectedChapter + "#page=")
	DownloadMangaChapter(baseURL, selectedChapter)
	sendMangaChapter(b, r, mangaList[mangaNumber].Title)
}

// GetMangaList - save all titles and links of a manga list.
func GetMangaList() {
	// TODO: remove number
	lastPage := CheckLastPageOfMangaList() / 10
	mangaList = make(map[int]mangaTitle)
	log.Println("Parsing manga list started!")
	offset := 0
	for i := 1; i <= lastPage; i++ {
		MangaListParser(offset)
		offset = offset + 70
	}
	log.Println("Parsing manga list finished!")
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

// DownloadMangaChapter - Download all chapter's pages.
func DownloadMangaChapter(URL string, selectedChapter string) {
	removeChaptersFolder()
	count, pre := CheckCountOfPages(URL, selectedChapter)
	log.Println("Starting download manga!")
	if pre == "#page=" {
		resp := getPageWithJS(URL + selectedChapter)
		doc, err := goquery.NewDocumentFromReader(resp.Body)
		checkError(err)
		mangaPageURL, ok := doc.Find("#fotocontext img").Attr("src")
		if ok {
			DowloadMangaPage(mangaPageURL)
		}
	}
	for i := 0; i < count; i++ {
		resp := getPageWithJS(URL + selectedChapter + pre + strconv.Itoa(i+1))
		doc, err := goquery.NewDocumentFromReader(resp.Body)
		checkError(err)
		mangaPageURL, ok := doc.Find("#fotocontext img").Attr("src")
		if ok {
			DowloadMangaPage(mangaPageURL)
		}
	}
	log.Println("Download finished!")
}

// CheckCountOfPages - Check count of pages.
func CheckCountOfPages(URL string, selectedChapter string) (c int, f string) {
	f = "#page="
	resp := getPageWithJS(URL + selectedChapter + f + strconv.Itoa(1))
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	checkError(err)
	c, err = strconv.Atoi(doc.Find(".top-block .pages-count").Text())
	if err != nil {
		f = "?mtr="
		resp := getPageWithJS(URL + selectedChapter + f + strconv.Itoa(1))
		doc, err := goquery.NewDocumentFromReader(resp.Body)
		checkError(err)
		log.Println(URL + selectedChapter + "?mtr=" + strconv.Itoa(1))
		c, err = strconv.Atoi(doc.Find(".top-block .pages-count").Text())
		checkError(err)
	}
	return
}

// DowloadMangaPage - Download one page.
func DowloadMangaPage(src string) {

	// Parse file name
	fileURL, err := url.Parse(src)
	checkError(err)
	path := fileURL.Path
	segments := strings.Split(path, "/")
	fileName := segments[len(segments)-1]

	// Create file with folder's path
	p, err := os.Create(filepath.Join("chapter", filepath.Base(fileName)))
	checkError(err)

	// Get file from URL
	filePage, err := httpClient().Get(src)
	checkError(err)
	defer filePage.Body.Close()

	//Put file in directory
	io.Copy(p, filePage.Body)
	defer p.Close()
}

func httpClient() *http.Client {
	client := http.Client{
		CheckRedirect: func(r *http.Request, via []*http.Request) error {
			r.URL.Opaque = r.URL.Path
			return nil
		},
	}

	return &client
}

func checkError(err error) {
	if err != nil {
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
	brower := NewPhantom()
	resp, err := brower.Download(p)
	checkError(err)
	return resp
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

func removeChaptersFolder() {
	os.RemoveAll("chapter/")
	os.MkdirAll("chapter/", 0777)
}

func getListFilesName() map[int]string {
	files, err := ioutil.ReadDir("chapter/")
	if err != nil {
		log.Fatal(err)
	}
	var fileNames = make(map[int]string)
	i := 0
	for _, f := range files {
		fileNames[i] = f.Name()
		i++
	}
	return fileNames
}

func main() {
	GetMangaList()
	b, err := tb.NewBot(tb.Settings{
		Token:  token,
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	})
	if err != nil {
		return
	}
	r := tb.Recipient(tb.ChatID(chatID))
	gocron.Every(10).Minutes().Do(downloadRandomMangaChapter, b, r)
	<-gocron.Start()
	b.Start()
}
