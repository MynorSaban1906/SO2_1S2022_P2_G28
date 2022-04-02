package main

// A simple example that shows how to send activity to Bubble Tea in real-time
// through a channel.29 Windows PowerShell
import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/gocolly/colly"
)

// variables globales
var page = ""
var monkey = 0
var wait = 0
var Nr = 0
var nFile = ""
var temporal = ""
var temporal2 responseMsg
var temporal3 data

type responseMsg struct {
	indice   int
	url      string
	estado   string
	palabras int
	enlaces  int
	cola     int
}

type data struct {
	Origin        string `json:"origin"`
	Couting_words int    `json:"couting_words"`
	Couting_links int    `json:"couting_links"`
	Sha           string `json:"sha"`
	Url           string `json:"url"`
	Monkey        string `json:"monkey"`
}

type Job_Mono struct {
	Url         string
	Busqueda    string
	Referencias int
}

type cache struct {
	mu    sync.RWMutex
	lista map[string]string
}

var Cola_Espera = &cache{lista: make(map[string]string)}

func Add_Mono(k string, v string) {
	if len(Cola_Espera.lista) < wait {
		Cola_Espera.mu.Lock()
		Cola_Espera.lista[k] = v
		Cola_Espera.mu.Unlock()
	}
}

func Saca_Cola(k string) {
	Cola_Espera.mu.Lock()
	delete(Cola_Espera.lista, k)
	Cola_Espera.mu.Unlock()
}

func L_referencia() string {
	Cola_Espera.mu.Lock()

	str := ""
	for k, v := range Cola_Espera.lista {
		str += fmt.Sprintf("%s -> %s \n", k, v)
	}
	Cola_Espera.mu.Unlock()
	return str

}

func Trabajo_Mono(sub chan responseMsg) tea.Cmd {

	return func() tea.Msg {
		jobs := make(chan Job_Mono, 100)
		results := make(chan Job_Mono, 100)

		for i := 0; i <= monkey; i++ {
			go mono(jobs, results, sub, i)
		}
		//go mono(jobs, results, sub, 0)
		//go mono(jobs, results, sub, 1)
		//go mono(jobs, results, sub, 2)

		jobs <- Job_Mono{page, "", Nr}
		for r := range results {
			x := r
			// fnt.Println(x.Busqueda)-12.01.22 pdt
			time.Sleep(time.Duration(1) * time.Second)
			Saca_Cola(x.Busqueda)
			jobs <- x
		}
		return nil

	}
}

func Espera_Mono(sub chan responseMsg) tea.Cmd {
	return func() tea.Msg {
		return responseMsg(<-sub)
	}
}

func mono(jobs <-chan Job_Mono, results chan<- Job_Mono, sub chan responseMsg, indice int) {


	for j := range jobs {

		Url := j.Url
		Nr := j.Referencias

		conteo_palabras := 0
		var enlaces []string
		var nombres_enlaces []string

		sub <- responseMsg{indice, Url, "trabajanding", 0, 0, -1}

		c := colly.NewCollector(colly.Async(false))
		c.OnRequest(func(r *colly.Request) {})

		c.OnHTML("div#mw-content-text p", func(e *colly.HTMLElement) {
			conteo_palabras += len(strings.Split(e.Text, " "))
			//fmt.Println("---------------------------------------------")
			//temporal = temporal + strconv.Itoa(conteo_palabras) + "\n"

			sub <- responseMsg{indice, Url, "trabajanding", conteo_palabras, len(enlaces), -1}
			temporal2.indice = indice
			temporal2.enlaces = len(enlaces)
			temporal2.palabras = conteo_palabras
			temporal2.url = Url

			temporal3.Origin = strconv.Itoa(indice)
			temporal3.Couting_links = len(enlaces)
			temporal3.Couting_words = conteo_palabras
			temporal3.Url = Url


			time.Sleep(500)
		})

		c.OnHTML("div#mw-content-text p a", func(e *colly.HTMLElement) {
			enlaces = append(enlaces, e.Request.AbsoluteURL(e.Attr("href")))
			nombres_enlaces = append(enlaces, e.Text)
			sub <- responseMsg{indice, Url, "trabajanding", conteo_palabras, len(enlaces), -1}
			temporal2.indice = indice
			temporal2.enlaces = len(enlaces)
			temporal2.palabras = conteo_palabras
			temporal2.url = Url

			temporal3.Origin = strconv.Itoa(indice)
			temporal3.Couting_links = len(enlaces)
			temporal3.Couting_words = conteo_palabras
			temporal3.Url = Url
		})

		c.OnScraped(func(e *colly.Response) {
			sub <- responseMsg{indice, Url, "descansanding", conteo_palabras, len(enlaces), -1}
			for i := 0; i < Nr; i++ {
				if len(enlaces) > i {
					aux := enlaces[i]
					nombre := nombres_enlaces[i]
					if len(results) < Nr {
						Add_Mono(nombre, aux)
						results <- Job_Mono{aux, nombre, Nr - 1}
					}
				}
			}
			time.Sleep(time.Duration(conteo_palabras/500) * time.Second)
		})
		c.Visit(Url)
		//fmt.Println(" ")
		//fmt.Println("---------------")
		//fmt.Println(temporal)
	}

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

func (m model) Init() tea.Cmd {
	return tea.Batch(
		spinner.Tick,
		Trabajo_Mono(m.sub),
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
			m.links = L_referencia()
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

	s := fmt.Sprintf(style.Render("Trabajando"))
	s += fmt.Sprintf("\n\n")

	for i := 0; i < monkey; i++ {
		s += fmt.Sprintf("%s %s url: %s \n palabras contandas : %d \n enlaces: %d \n\n", m.monos[i], m.estados[i], m.urls[i], m.palabras[i], m.enlaces[i])
	}

	s += fmt.Sprintf("En espera")
	s += fmt.Sprintf("\n\n %s ", m.links)

	//s += fmt.Sprintf(" ---------------------------------------------------\n")
	//s += fmt.Sprintf("%+v", temporal2)

	//convierte el resultado en bytes
	b, _ := json.MarshalIndent(temporal3, "", "    ")
	//fmt.Printf("%s", b)
	temporal = temporal + string(b)
	c := []byte(temporal)
	//se crea el archivo donde se guarda todo
	file, _ := os.Create(nFile + ".json")

	defer file.Close()

	//se escribe resultados
	file.Write(c)

	s += fmt.Sprintf("\n\n presione para salir ")

	if m.quitting {
		s += "\n"
	}
	return s

}

func main() {
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
	//var prueba [2000]string

	//fmt.Println(monkey)
	//for i := 0; i < monkey; i++ {
	//	fmt.Println(i)

	//	prueba[i] = "id" + strconv.Itoa(i)
	//}

	p := tea.NewProgram(model{
		sub:      make(chan responseMsg),
		monos:    []string{"id_1", "id_2", "id_3", "id_4", "id_5", "id_6", "id_7", "id_8", "id_9", "id_10", "id_", "id_", "id_", "id_", "id_", "id_", "id_", "id_", "id_", "id_", "id_", "id_", "id_", "id_", "id_", "id_", "id_", "id_", "id_", "id_"},
		urls:     []string{"", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", ""},
		estados:  []string{"", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", ""},
		palabras: []int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		enlaces:  []int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},

		spinner: spinner.New(),
	})

	if p.Start() != nil {
		fmt.Println("could not start program")
		os.Exit(1)
	}
}
