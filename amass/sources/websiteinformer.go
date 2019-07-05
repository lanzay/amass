// Copyright 2019 Lanzay. All rights reserved.
// Use of this source code is governed by Apache 2 LICENSE that can be found in the LICENSE file.

package sources

import (
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"log"
	"strings"
	"time"

	"github.com/lanzay/Amass/amass/core"
	"github.com/lanzay/Amass/amass/utils"
)

type WebsiteInformer struct {
	core.BaseService

	SourceType string
}

// WebsiteInformer returns he object initialized, but not yet started.
func NewWebsiteInformer(config *core.Config, bus *core.EventBus) *WebsiteInformer {
	w := &WebsiteInformer{
		SourceType: core.SCRAPE,
	}

	w.BaseService = *core.NewBaseService(w, "Website Informer", config, bus)
	return w
}

// OnStart implements the Service interface
func (w *WebsiteInformer) OnStart() error {
	w.BaseService.OnStart()

	go w.processRequests()
	return nil
}

func (w *WebsiteInformer) processRequests() {
	for {
		select {
		case <-w.Quit():
			return
		case req := <-w.DNSRequestChan():
			if w.Config().IsDomainInScope(req.Domain) {
				w.executeQuery(req.Domain)
			}
		case <-w.AddrRequestChan():
		case <-w.ASNRequestChan():
		case <-w.WhoisRequestChan():
		}
	}
}

func (w *WebsiteInformer) executeQuery(domain string) {
	re := w.Config().DomainRegex(domain)
	if re == nil {
		return
	}

	t := time.NewTicker(time.Second)
	defer t.Stop()

	for {
		select {
		case <-w.Quit():
			return
		case <-t.C:
			domains, err := w.domainsByMail(domain)
			if err != nil {
				w.Config().Log.Printf("%s: %s: %v", w.String(), domain, err)
				return
			}

			for _, domain := range domains {
				w.Bus().Publish(core.NewNameTopic, &core.DNSRequest{
					Name:   cleanName(domain),
					Domain: domain,
					Tag:    w.SourceType,
					Source: w.String(),
				})
			}
		}
	}
}

// https://website.informer.com/att.com/emails
// div.list-email:nth-child(1)>ul>li>a

// https://website.informer.com/email/att-domains@att.com
// div.list-sites-e>div>div>a

// #show-more
// https://website.informer.com/ajax/email/sites?email_id=309&sort_by=popularity&skip=357,710,2177,95245403,60296807,54312855,86338,59705233,8524289,129803433&max_index=10
// https://website.informer.com/ajax/email/sites?email_id=309

//page, err := utils.RequestWebPage(u, nil, nil, "", "")
func (w *WebsiteInformer) domainsByMail(domain string) ([]string, error) {

	u := fmt.Sprintf("https://website.informer.com/%s/emails", domain)
	page, err := utils.RequestWebPage(u, nil, nil, "", "")
	if err != nil {
		return nil, err
	}
	r := strings.NewReader(page)
	doc, err := goquery.NewDocumentFromReader(r)
	var mails []string
	doc.Find("div.list-email:nth-child(1)>ul>li>a").Each(func(i int, s *goquery.Selection) {
		if mail, ok := s.Attr("href"); ok {
			mails = append(mails, mail)
		}
	})

	var domains []string
	var ids []string
	var mailId string
	for _, mail := range mails {
		u := fmt.Sprintf("https://website.informer.com%s", mail)
		page, err := utils.RequestWebPage(u, nil, nil, "", "")
		if err != nil {
			return nil, err
		}
		r := strings.NewReader(page)
		doc, err := goquery.NewDocumentFromReader(r)
		mailId, _ = doc.Find("#show-more").Attr("email-id")
		doc.Find("div.list-sites-e>div>div>a").Each(func(i int, s *goquery.Selection) {
			if domain, ok := s.Attr("href"); ok {
				domains = append(domains, strings.TrimPrefix(domain, "/"))
				if id, ok := s.Parent().Parent().Attr("site-id"); ok {
					ids = append(ids, id)
				}
			}
		})
	}

	type respW struct {
		Sites    string `json:"sites"`
		ShowMore bool   `json:"showMore"`
	}
	i := 0
	part := &respW{ShowMore: true}
	for part.ShowMore {
		i += 10
		u = fmt.Sprintf("https://website.informer.com/ajax/email/sites?email_id=%s&sort_by=popularity&skip=%s&max_index=%d", mailId, strings.Join(ids, ","), i)
		page, err = utils.RequestWebPage(u, nil, nil, "", "")
		//println(page)
		checkErr(err)
		err = json.Unmarshal([]byte(page), part)
		checkErr(err)
		r = strings.NewReader(part.Sites)
		doc, err = goquery.NewDocumentFromReader(r)
		checkErr(err)
		doc.Find("div.one-sites-e>div.left-sites-e>a").Each(func(i int, s *goquery.Selection) {
			if domain, ok := s.Attr("href"); ok {
				domains = append(domains, strings.TrimPrefix(domain, "/"))
				if id, ok := s.Parent().Parent().Attr("site-id"); ok {
					ids = append(ids, id)
				}
			}
		})
	}

	return domains, nil

}
func checkErr(err error) {
	if err != nil {
		log.Println("[E]", err)
	}
}
