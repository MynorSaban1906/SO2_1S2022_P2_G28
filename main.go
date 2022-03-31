package main

import (
	//"crypto/sha1"
	//"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	//"regexp"
	"strings"
	"github.com/gocolly/colly"
	"sync"
	"time"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"strconv"
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


type model struct {
	sub      chan responseMsg
	monos    []string
	urls     []string
	palabras []int
	enlaces  []int
	estados  []string
	links    string
	rola     int
	spinner  spinner.Model
	quitting bool
}

//structura que maneja los links 
type Trabajo_mono struct {
	url string 
	busqueda string 
	refrerencias int 
	tam_cola int 
}

type Cache_Urls struct {
	mu    sync.RWMutex
	lista map[string]string

}


type responseMsg struct {
	indice   int
	url      string
	estado   string
	palabras int
	enlaces  int
	cola     int
}

var Cola_espera = &Cache_Urls{lista: make(map[string]string)}


func agregar_accion(Pos string, Valor string){
	if len(Cola_espera.lista) < wait {
		Cola_espera.mu.Lock()
		Cola_espera.lista[Pos] = Valor
		Cola_espera.mu.Unlock()
	}

}

func quitar_accion(valor string) {
	Cola_espera.mu.Lock()
	delete(Cola_espera.lista, valor)
	Cola_espera.mu.Unlock()
}

func leer_referencia() string {
	Cola_espera.mu.Lock()
	str := ""
	for k, v := range Cola_espera.lista {
		str += fmt.Sprintf("%s -> %s \n", k, v)
	}
	Cola_espera.mu.Unlock()
	return str

}




// variables globales
var page = ""
var monkey = 0
var wait = 0
var Nr = 0
var nFile = ""
var Tam_Cola = 0
var resultados data
var script = "{"


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


	p := tea.NewProgram(model{
		sub:      make(chan responseMsg),
		monos:    []string{"", "", "","", "","","", "","","", "","","","","","","",""},
		urls:     []string{"", "", "","", "", "","", "", "","", "", "","", "", "","", "", ""},
		estados:  []string{"", "", "","", "", "","", "", "","", "", "","", "", "","", "", ""},
		palabras: []int{0, 0, 0,0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,0,0},
		enlaces:  []int{0, 0, 0,0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,0,0},

		spinner: spinner.New(),
	})

	if p.Start() != nil {
		fmt.Println("could not start program")
		os.Exit(1)
	}

	
	//convierte el resultado en bytes
	fmt.Println(resultados)

	b, _ := json.MarshalIndent(resultados, "", "    ")
	fmt.Printf("%s", b)
	//se crea el archivo donde se guarda todo
	file, _ := os.Create(nFile + ".json")

	defer file.Close()
	//se escribe resultados
	file.Write(b)

}


//-------------------------
func Accion_Mono(sub chan responseMsg) tea.Cmd {


	return func() tea.Msg {
	jobs := make(chan Trabajo_mono, 100)
	results := make(chan Trabajo_mono, 100)


	for i := 0; i <= monkey; i++ {
		go mono(jobs, results, sub, i)	

	}
	

	jobs <- Trabajo_mono{page, "Chuck", Nr, wait}
	for r := range results {
		x := r
		// fnt.Println(x.Busqueda)-12.01.22 pdt
		time.Sleep(time.Duration(1) * time.Second)
		quitar_accion(x.busqueda)
		jobs <- x
	}
	return nil
	}

}


func Espera_Mono(sub chan responseMsg) tea.Cmd{
	return func() tea.Msg {
		return responseMsg(<-sub)
	}
	
}



func mono(jobs chan Trabajo_mono, results chan Trabajo_mono, sub chan responseMsg, indice int) data {
	
	for j := range jobs {
		Url := j.url
		Nr := j.refrerencias

		conteo_palabras := 0
		var enlaces []string
		var nombres_enlaces []string

		sub <- responseMsg{indice, Url, "Trabajando", 0, 0, -1}

		c := colly.NewCollector(colly.Async(false))
		c.OnRequest(func(r *colly.Request) {})

		c.OnHTML("div#mw-content-text p", func(e *colly.HTMLElement) {
			conteo_palabras += len(strings.Split(e.Text, " "))
			sub <- responseMsg{indice, Url, "Trabajando", conteo_palabras, len(enlaces), -1}

			resultados.Couting_links = len(enlaces)
			resultados.Url = Url
			resultados.Monkey = "Mono_" + strconv.Itoa(indice)
			resultados.Couting_words =  conteo_palabras
			

			time.Sleep(500)
		})

		c.OnHTML("div#mw-content-text p a", func(e *colly.HTMLElement) {
			enlaces = append(enlaces, e.Request.AbsoluteURL(e.Attr("href")))
			nombres_enlaces = append(enlaces, e.Text)
			sub <- responseMsg{indice, Url, "Trabajando", conteo_palabras, len(enlaces), -1}
			
			resultados.Couting_links = len(enlaces)
			resultados.Url = Url
			resultados.Monkey = "Mono_" + strconv.Itoa(indice)
			resultados.Couting_words =  conteo_palabras

		})

		c.OnScraped(func(e *colly.Response) {
			sub <- responseMsg{indice, Url, "Desacansando", conteo_palabras, len(enlaces), -1}
			for i := 0; i < Nr; i++ {
				if len(enlaces) > i {
					aux := enlaces[i]
					nombre := nombres_enlaces[i]
					if len(results) < Nr {
						agregar_accion(nombre, aux)
						results <- Trabajo_mono{aux, nombre, Nr - 1, wait}
						
						resultados.Couting_links = len(enlaces)
						resultados.Url = Url
						resultados.Monkey = "Mono_" + strconv.Itoa(indice)
						resultados.Couting_words =  conteo_palabras
			
					}
				}
			}
			time.Sleep(time.Duration(conteo_palabras/500) * time.Second)
		})

		//sha := sha1.New()
		//sha.Write([]byte(pharagraph))


		c.Visit(Url)
		return resultados
	}	
	return resultados
}


func (m model) Init() tea.Cmd {
	return tea.Batch(
		spinner.Tick,
		Accion_Mono(m.sub),
		Espera_Mono(m.sub),
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case tea.KeyMsg:
		m.quitting = true
		return m, tea.Quit

	case responseMsg:
		respuesta := msg.(responseMsg)
		if respuesta.cola == -1 {
			m.urls[respuesta.indice] = respuesta.url
			m.palabras[respuesta.indice] = respuesta.palabras
			m.enlaces[respuesta.indice] = respuesta.enlaces
			m.estados[respuesta.indice] = respuesta.estado
			m.links = leer_referencia()
		}

		return m, Espera_Mono(m.sub) // wait for nest envent

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	default:
		return m, nil

	}
}

func (m model) View() string {
	var style = lipgloss.NewStyle().
		Bold(true).
		Background(lipgloss.Color("#77D56F4")).
		Height(1)

	s := fmt.Sprintf(style.Render(" --- silencio nonos trabajando  --"))
	s += fmt.Sprintf("\n\n")

	for i := 0; i < monkey; i++ {
		s += fmt.Sprintf("%s %s url: %s \n palabras contandas : %d \n enlaces: %d \n\n", m.monos[i], m.estados[i], m.urls[i], m.palabras[i], m.enlaces[i])
	}

	s += fmt.Sprintf(" ---- cola de trabajo   ---")
	s += fmt.Sprintf("\n\n %s ", m.links)
	s += fmt.Sprintf("\n\n presione para salir ")

	if m.quitting {
		s += "\n"
	}
	return s

}


/*
///------------------
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

		//parrafo a sh256
		sha := sha1.New()
		sha.Write([]byte(pharagraph))

		results.Sha = hex.EncodeToString(sha.Sum(nil))

		results.Couting_words = words
		results.Couting_links = linkcount
	})

	results.Url = Url
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
*/
