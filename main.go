package main

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/gocolly/colly"
)

type data struct {
	origin        string
	couting_words int
	couting_links int
	sha           string
	url           string
	mono          string
}

func main() {

	result := scraper("https://es.m.wikipedia.org/wiki/Red_de_computadoras")

	fmt.Printf("%+v", result)

}

func scraper(page string) data {
	//declara la estructura data que contendra los datos de la pagina visitada
	var results data
	linkcount := 0
	c := colly.NewCollector(colly.AllowedDomains("https://es.m.wikipedia.org", "es.m.wikipedia.org"))
	pharagraph := ""
	c.OnHTML("div.mw-parser-output", func(first *colly.HTMLElement) {
		//ciclo para obtener todos los parrafos que estan en las secciones
		first.ForEach("p", func(i int, second *colly.HTMLElement) {
			//ciclo para obeter los links
			second.ForEach("a[href]", func(i int, third *colly.HTMLElement) {
				link := third.Attr("href")

				// se quita los links que inician con #
				if !strings.ContainsAny(link, "#") {
					linkcount++
					//obtiene los links para hacer las demas busquedas
					fmt.Printf("Link: %s -> %s\n", third.Text, link)
				}

			})

			//junta los parrafos de todas las secciones
			pharagraph += second.Text
		})

		//cuenta el numero de palabras que hay en la pagina en las etiquetas <p>
		words := wordCount(pharagraph)
		results.couting_words = words
		results.couting_links = linkcount
	})

	results.url = page
	c.Visit(page)

	return results

}

func wordCount(value string) int {
	// Match no contando espacios solo caracteres juntos
	re := regexp.MustCompile(`[\S]+`)

	// busca todos los que hacen march y retorna el valor
	results := re.FindAllString(value, -1)
	return len(results)
}
