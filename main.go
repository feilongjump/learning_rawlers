package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gocolly/colly/v2"
)

// CrawlerItem 页面数据
type CrawlerItem struct {
	Link           string
	Title          string
	ProductDetails bool
}

var CrawlerData = make(map[string]CrawlerItem)

func main() {
	// 访问的链接
	url := ""

	c := colly.NewCollector()

	// 爬取页面所有产品链接
	c.OnHTML("#unit-bWp9AvDc07 a.unit-list__image[href]", func(e *colly.HTMLElement) {
		e.Request.Visit(e.Attr("href"))
	})

	visitCount := 1
	visitMaxCount := 2
	c.OnHTML(
		"li.base-pagination__item.base-pagination__item--next.page-next > a.base-pagination__link",
		func(e *colly.HTMLElement) {
			link := e.Attr("href")
			if link != "javascript:;" && visitCount <= visitMaxCount {
				e.Request.Visit(link)

				fmt.Println(visitCount)
			}
		},
	)

	// 赋值产品标题
	setTitle(c)

	// 赋值产品详情
	setProductDetails(c)

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Progress:", r.URL)
		customURL := "/product_list" + strconv.Itoa(visitCount) + ".html"
		if r.URL.Path == customURL || r.URL.Path == "/product.html" {
			visitCount++
			fmt.Println("============= next page :", visitCount, "=============")
			fmt.Println("============= 以下为当前页（", visitCount-1, "）的请求链接 =============")
		}
		// 赋值产品链接
		link := r.URL.String()
		if _, ok := CrawlerData[link]; !ok {
			CrawlerData[link] = CrawlerItem{
				Link: link,
			}
		}
	})
	c.Visit(url)

	// fmt.Println(CrawlerData)
}

// setTitle 赋值产品标题
func setTitle(c *colly.Collector) {

	c.OnHTML("h1.unit-detail_title.nostyle", func(e *colly.HTMLElement) {
		// 获取产品标题
		link := e.Request.URL.String()
		if item, ok := CrawlerData[link]; ok {
			item.Title = e.Text
			CrawlerData[link] = item
		} else {
			CrawlerData[link] = CrawlerItem{
				Link:  link,
				Title: e.Text,
			}
		}
	})
}

// setProductDetails 赋值产品详情
func setProductDetails(c *colly.Collector) {

	c.OnHTML(".unit-detail-html-tabs__nav-box a.unit-detail-html-tabs__nav-link.nav-link.active", func(e *colly.HTMLElement) {
		// 获取产品详情是否存在
		link := e.Request.URL.String()
		if item, ok := CrawlerData[link]; ok {
			item.ProductDetails = strings.TrimSpace(e.Text) == ""
			CrawlerData[link] = item
		} else {
			CrawlerData[link] = CrawlerItem{
				Link:           link,
				ProductDetails: strings.TrimSpace(e.Text) == "",
			}
		}
	})
}
