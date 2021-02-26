package main

import (
	"bufio"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

//Page is used for template
type Page struct { //Creates the Page structure which allows the personalization of the website "/ascii-art"
	ColorTxt string //value used to add color selection
	FontSize string //value used to add a size to the Ascii Generator
	ColorBG  string //value used to add color to the background of the website "/ascii-art"
}

//ReadFile returns an array of string which is the same as the file (line = line)
func ReadFile(StylizedFile string) []string {
	var source []string
	file, _ := os.Open(StylizedFile)  // opens the .txt
	scanner := bufio.NewScanner(file) // scanner scans the file
	scanner.Split(bufio.ScanLines)    // sets-up scanner preference to read the file line-by-line
	for scanner.Scan() {              // loop that performs a line-by-line scan on each new iteration
		if scanner.Text() != "" {
			source = append(source, scanner.Text()) // adds the value of scanner (that contains the characters from StylizedFile) to source
		}
	}
	file.Close() // closes the file
	return source
}

//BuildTemplate edit the configuration of the ascii-art page
func BuildTemplate(w http.ResponseWriter, r *http.Request, ColorTxt string, FontSize string, ColorBG string) {
	p := Page{ColorTxt, FontSize, ColorBG}                               // Associates the elements of ascii-art.html with main.go
	parsedTemplate, _ := template.ParseFiles("templates/ascii-art.html") // States the files that requires a modification

	parsedTemplate.Execute(w, p) // Executes the modification
	// The variable p will be depicted as "." inside the layout
	// Exemple : {{.}} == p
}

func pageChecker(w http.ResponseWriter, r *http.Request, page string) bool {
	if r.URL.Path != page { // Checks if we are in  /ascii-art
		http.Error(w, "404 not found.", http.StatusNotFound) // Sends a 404 error (page not found)
		return true
	}
	return false
}

func formGetter(w http.ResponseWriter, r *http.Request) (FontSize, ColorBG, ColorTxt, Fontlist, Text string) {
	if err := r.ParseForm(); err != nil { //If an error occurs during the POST request
		fmt.Fprintf(w, "Can't get input data")
		fmt.Fprintf(w, "ParseForm() err: %v", err)
		return
	}
	// Saves the value from the form tag of the root page ("index.html")
	FontSize = r.FormValue("FontSize")
	ColorBG = r.FormValue("ColorBG")
	ColorTxt = r.FormValue("ColorList")
	Fontlist = r.FormValue("Fontlist")
	Text = r.FormValue("Text")
	return
}

//Provide support for /ascii-art page
func asciiHandler(w http.ResponseWriter, r *http.Request) {
	if pageChecker(w, r, "/ascii-art") {
		return
	}

	fmt.Println("pass")

	FontSize, ColorBG, ColorTxt, Fontlist, Text := formGetter(w, r)
	if Fontlist == "" {
		return
	}

	if Text == "" { // Tells to the user that he didn't follow the instructions and quit
		fmt.Fprintf(w, "	Please enter the text in the right position then hit \"Go convert\" button")
		return
	}

	arguments := []rune(Text)
	for index := range arguments {
		if arguments[index] < 32 || arguments[index] > 126 { // Conditions to check if there are non-printable characters
			if arguments[index] != '\n' {
				if arguments[index] != '\r' {
					fmt.Fprintf(w, "	You wrote a non-printable character... Please, try again")
					return
				}
			}
		}
	}
	txtfile := ReadFile(Fontlist) // Recovers the right txt file

	BuildTemplate(w, r, ColorTxt, FontSize, ColorBG) // Calls the configuration of the webpage

	var start int
	w.Write([]byte("<pre>")) // Uses "pre"(html tag) to handle breakline
	for index := range arguments {
		if arguments[index] == '\r' {
			for _, line := range asciiGenerator(arguments[start:index], txtfile) {
				fmt.Fprintf(w, line)
			}
			fmt.Fprintln(w)
			start = index + 2
		} else if index == len(arguments)-1 {
			for _, line := range asciiGenerator(arguments[start:], txtfile) {
				fmt.Fprintf(w, line)
			}
		}
	}
	w.Write([]byte("</pre>"))
}

//asciiGenerator prints the stylized characters
func asciiGenerator(arguments []rune, txtfile []string) []string {
	var push []string
	for ligne := 0; ligne < 8; ligne++ { // Each character is composed of 8 lines
		for index, char := range arguments {
			push = append(push, txtfile[ligne+(int(char)-32)*8])
			if index == len(arguments)-1 && ligne != 7 { // Jumps a newline when it is required
				push = append(push, "\n")
				break
			}
		}
	}
	return push
}

//fileDowload allow the user to download file from the server
func fileDownload(w http.ResponseWriter, r *http.Request, Filename string) {
	fmt.Println("Client requests: " + Filename)

	Openfile, _ := os.Open(Filename)
	defer Openfile.Close() //Close after function return

	//File is found, create and send the correct headers

	//Get the Content-Type of the file
	//Create a buffer to store the header of the file in
	FileHeader := make([]byte, 512)
	//Copy the headers into the FileHeader buffer
	Openfile.Read(FileHeader)
	//Get content type of file
	FileContentType := http.DetectContentType(FileHeader)

	//Get the file size
	FileStat, _ := Openfile.Stat()                     //Get info from file
	FileSize := strconv.FormatInt(FileStat.Size(), 10) //Get file size as a string

	//Send the headers
	w.Header().Set("Content-Disposition", "attachment; filename="+Filename)
	w.Header().Set("Content-Type", FileContentType)
	w.Header().Set("Content-Length", FileSize)

	//Send the file
	//We read 512 bytes from the file already, so we reset the offset back to 0
	Openfile.Seek(0, 0)
	io.Copy(w, Openfile) //'Copy' the file to the client
	return
}

//exportHandler handle exportation system
func exportHandler(w http.ResponseWriter, r *http.Request) {
	if pageChecker(w, r, "/export") {
		return
	}

	_, _, _, Fontlist, Text := formGetter(w, r) // For now we don't need more than Text and Fontlist
	if Fontlist == "" {
		return
	}

	format := r.URL.Query().Get("format")

	if Text == "" { // Tells to the user that he didn't follow the instructions and quit
		fmt.Fprintf(w, "	Please enter the text in the right position then hit \"Export to %s\" button", strings.ToUpper(format))
		return
	}

	if format == "txt" {
		exportTXT(Text, Fontlist)
	} else {
		fmt.Println("Not yet supported")
		return
	}
	fileDownload(w, r, "export."+format)
}

//exportTXT create a .txt file and put ascii-art inside
func exportTXT(Text string, Fontlist string) {
	file, err := os.Create("./export.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	txtfile := ReadFile(Fontlist) // Recovers the right txt file

	arguments := []rune(Text)

	var start int
	for index := range arguments {
		if arguments[index] == '\r' {
			for _, line := range asciiGenerator(arguments[start:index], txtfile) {
				file.WriteString(line)
			}
			file.WriteString("\n")
			start = index + 2
		} else if index == len(arguments)-1 {
			for _, line := range asciiGenerator(arguments[start:], txtfile) {
				file.WriteString(line)
			}
		}
	}
}

func main() {
	http.Handle("/", http.FileServer(http.Dir("./templates"))) // handle root (main) webpage
	http.HandleFunc("/ascii-art", asciiHandler)                // handle ascii-art page
	http.HandleFunc("/export", exportHandler)                  // handle generator page

	fmt.Printf("Starting server at port 8080\n")
	fmt.Println("Go on http://localhost:8080") // Prints the link of the website on the command prompt
	fmt.Printf("\nTo shutdown the server and exit the code hit \"crtl+C\"\n")
	if err := http.ListenAndServe(":8080", nil); err != nil { // Launches the server on port 8080 if port 8080 is not already busy, else quit
		log.Fatal(err)
	}
}
