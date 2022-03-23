package main

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/gocolly/colly"
)

//structura de json
type data struct {
	Origin        string `json:"origin"`
	Couting_words int    `json:"couting_words"`
	Couting_links int    `json:"couting_links"`
	Sha           string `json:"sha"`
	Url           string `json:"url"`
	Monkey        string `json:"monkey"`
}

// variables globales
var page = ""
var monkey = 0
var wait = 0
var Nr = 0
var nFile = ""

func main() {
	// pagina de pruebas  https://es.wikipedia.org/wiki/Red_de_computadoras

	fmt.Println("---------------------------------------------")
	fmt.Print("Cantidad de monos buscadores : ")
	fmt.Scan(&monkey)
	fmt.Print("Tamaño de la cola de espera : ")
	fmt.Scan(&wait)
	fmt.Print("numero de Referencias : ")
	fmt.Scan(&Nr)
	fmt.Print("URL de la Pagina : ")
	fmt.Scan(&page)
	fmt.Print("Nombre del archivo : ")
	fmt.Scan(&nFile)
	fmt.Println("---------------------------------------------")
	//resultado del scraper
	result := scraper(page)
	//convierte el resultado en bytes
	b, _ := json.MarshalIndent(result, "", "    ")
	fmt.Printf("%s", b)
	//se crea el archivo donde se guarda todo
	file, _ := os.Create(nFile + ".json")

	defer file.Close()

	//se escribe resultados
	file.Write(b)

}

func scraper(page string) data {
	//declara la estructura data que contendra los datos de la pagina visitada
	var results data
	linkcount := 0
	c := colly.NewCollector()
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
		results.Couting_words = words
		results.Couting_links = linkcount
	})

	results.Url = page
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
