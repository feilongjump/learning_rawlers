package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/gocolly/colly/v2"
	"github.com/xuri/excelize/v2"
)

// CrawlerItem 页面数据
type CrawlerItem struct {
	Link           string
	Title          string
	ProductDetails bool
}

var CrawlerData = make(map[string]CrawlerItem)

func main() {
	// 页面的链接
	url := ""
	if err := crawler(url); err != nil {
		fmt.Println(err.Error())
	} else {
		saveExcel()
	}
}

// crawler 爬取页面
func crawler(url string) error {

	c := colly.NewCollector()

	// 爬取页面所有产品链接
	c.OnHTML("#unit-bWp9AvDc07 a.unit-list__image[href]", func(e *colly.HTMLElement) {
		e.Request.Visit(e.Attr("href"))
	})

	// 限制数量访问
	visitCount := 1
	visitMaxCount := 688
	c.OnHTML(
		"li.base-pagination__item.base-pagination__item--next.page-next > a.base-pagination__link",
		func(e *colly.HTMLElement) {
			link := e.Attr("href")
			if link != "javascript:;" && visitCount <= visitMaxCount {
				e.Request.Visit(link)
			}
		},
	)

	// 赋值产品标题
	setTitle(c)

	// 赋值产品详情
	setProductDetails(c)

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Progress:", r.URL)
		path := r.URL.Path
		customURL := "/product_list" + strconv.Itoa(visitCount) + ".html"
		if path == customURL || path == "/product.html" {
			visitCount++
			fmt.Println("============= next page :", visitCount, "=============")
			fmt.Println("============= 以下为当前页（", visitCount-1, "）的请求链接 =============")
		} else {

			// 赋值产品链接
			link := r.URL.String()
			if _, ok := CrawlerData[link]; !ok {
				CrawlerData[link] = CrawlerItem{
					Link: link,
				}
			}
		}
	})
	if err := c.Visit(url); err != nil {
		log.Fatal(err)
		return err
	}

	return nil
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
			item.ProductDetails = strings.TrimSpace(e.Text) != ""
			CrawlerData[link] = item
		} else {
			CrawlerData[link] = CrawlerItem{
				Link:           link,
				ProductDetails: strings.TrimSpace(e.Text) != "",
			}
		}
	})
}

// saveExcel 保存到 excel 中
func saveExcel() {
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	}()
	// 创建一个工作表
	index, err := f.NewSheet("Sheet1")
	if err != nil {
		fmt.Println(err)
		return
	}
	// 设置列宽
	f.SetColWidth("Sheet1", "A", "B", 120)
	f.SetColWidth("Sheet1", "C", "C", 20)
	// 设置单元格的值
	cellNum := 2
	f.SetCellValue("Sheet1", "A1", "产品链接")
	f.SetCellValue("Sheet1", "B1", "产品标题")
	f.SetCellValue("Sheet1", "C1", "是否存在产品详情")

	for _, v := range CrawlerData {
		cellNumStr := strconv.Itoa(cellNum)

		f.SetCellValue("Sheet1", "A"+cellNumStr, strings.TrimSpace(v.Link))
		f.SetCellValue("Sheet1", "B"+cellNumStr, strings.TrimSpace(v.Title))
		f.SetCellValue("Sheet1", "C"+cellNumStr, v.ProductDetails)

		cellNum++
	}
	// 设置工作簿的默认工作表
	f.SetActiveSheet(index)
	// 根据指定路径保存文件
	if err := f.SaveAs("product.xlsx"); err != nil {
		fmt.Println(err)
	}
}
