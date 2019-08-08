// Copyright 2019 Lanzay. All rights reserved.
// Use of this source code is governed by Apache 2 LICENSE that can be found in the LICENSE file.

package sources

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/lanzay/Amass/amass/core"
	"github.com/lanzay/Amass/amass/utils"
	"log"
	"strings"
	"time"
)

var (
	blackListMails = []string{
		"do-not-use@privacy-protected.email",
	}
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

	w.SetActive()

	t := time.NewTicker(time.Second)
	defer t.Stop()

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
			if w.Config().IncludeDomainByAdminMail {
				w.Config().AddDomain(cleanName(domain)) //TODO !!! ADD NEW domain BY Admin eMail
			}
			w.Bus().Publish(core.NewNameTopic, &core.DNSRequest{
				Name:   cleanName(domain),
				Domain: domain,
				Tag:    w.SourceType,
				Source: w.String(),
			})
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

	blackList := []string{
		"do-not-use@privacy-protected.email",
	}

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

		for _, bl := range blackList {
			if strings.EqualFold(bl, mail) {
				log.Println("[W] eMail in black list", domain, mail)
				err = errors.New("eMail in black list")
				break
			}
		}

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
		if err != nil {
			break
		}
		r = strings.NewReader(part.Sites)
		doc, err = goquery.NewDocumentFromReader(r)
		checkErr(err)
		if err != nil {
			break
		}
		doc.Find("div.one-sites-e>div.left-sites-e>a").Each(func(i int, s *goquery.Selection) {
			if domain, ok := s.Attr("href"); ok {
				domains = append(domains, strings.TrimPrefix(domain, "/"))
				if id, ok := s.Parent().Parent().Attr("site-id"); ok {
					ids = append(ids, id)
				}
			}
		})
	}

	return domains, err

}
func checkErr(err error) {
	if err != nil {
		log.Println("[E]", err)
	}
}

//============================
func (w *WebsiteInformer) GetDomainsByDomain(domain string) ([]string, []string, error) {
	mails, err := w.GetDomainMails(domain)
	domains, err := w.GetDomainsByMails(mails)
	return domains, mails, err
}

//Получаем eMails домена из WhoIs и т.п.
func (w *WebsiteInformer) GetDomainMails(domain string) ([]string, error) {

	u := fmt.Sprintf("https://website.informer.com/%s/emails", domain)
	page, err := utils.RequestWebPage(u, nil, nil, "", "")
	if err != nil {
		return nil, err
	}
	r := strings.NewReader(page)
	doc, err := goquery.NewDocumentFromReader(r)
	if utils.CheckErr(err) {
		return nil, err
	}
	var mails []string
	doc.Find("div.list-email:nth-child(1)>ul>li>a").Each(func(i int, s *goquery.Selection) {
		if mail, ok := s.Attr("href"); ok {
			mail = strings.TrimPrefix(mail, "/email/")
			mails = append(mails, mail)
		}
	})
	return mails, nil
}

//Получаем все домены у которых при регистрации указаны eMails
func (w *WebsiteInformer) GetDomainsByMails(mails []string) ([]string, error) {

	var err error
	var domains []string
	var ids []string
	var mailId string

	for _, mail := range mails {
		if mailInBlackList(mail) {
			//w.Config().Log.Printf("[W] eMail in black list", mail)
			continue
		}

		u := fmt.Sprintf("https://website.informer.com/email/%s", mail)
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

	//Scroll all resp 10...
	type respW struct {
		Sites    string `json:"sites"`
		ShowMore bool   `json:"showMore"`
	}
	i := 0
	page := ""
	part := &respW{ShowMore: true}
	for part.ShowMore {
		i += 10
		u := fmt.Sprintf("https://website.informer.com/ajax/email/sites?email_id=%s&sort_by=popularity&skip=%s&max_index=%d", mailId, strings.Join(ids, ","), i)
		page, err = utils.RequestWebPage(u, nil, nil, "", "")
		if utils.ErrLog(err) {
			break
		}
		err = json.Unmarshal([]byte(page), part)
		if utils.ErrLog(err) {
			break
		}

		r := strings.NewReader(part.Sites)
		doc, err := goquery.NewDocumentFromReader(r)
		utils.CheckErr(err)
		if err != nil {
			break
		}
		doc.Find("div.one-sites-e>div.left-sites-e>a").Each(func(i int, s *goquery.Selection) {
			if domain, ok := s.Attr("href"); ok {
				domains = append(domains, strings.TrimPrefix(domain, "/"))
				if id, ok := s.Parent().Parent().Attr("site-id"); ok {
					ids = append(ids, id)
				}
			}
		})
	}
	return domains, err
}

func mailInBlackList(mail string) bool {

	for _, bl := range blackListMails {
		if strings.EqualFold(bl, mail) {
			return true
		}
	}
	return false
}
