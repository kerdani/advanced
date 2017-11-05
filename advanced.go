package main
import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strconv"
	//"strings"
	"time"
  "os"



	//"github.com/gorilla/mux"
	"io/ioutil"

	cors "github.com/heppu/simple-cors"
)

var places space

// type returnObject struct {
// 	From  string  `json:"From"`
// 	To    string  `json:"To"`
// 	Value float64 `json:"Value"`
// }

type space struct {
	Results []struct {
		Formatted string `json:"formatted_address"`

		Icon string

		Name string

		Rating float64
	}
}


var (
	// WelcomeMessage A constant to hold the welcome message
	WelcomeMessage = "Hi, where do you want to go?"

	// sessions = {
	//   "uuid1" = Session{...},
	//   ...
	// }
	sessions = map[string]Session{}

	processor = sampleProcessor
)

type (
	// Session Holds info about a session
	Session map[string]interface{}

	// JSON Holds a JSON object
	JSON map[string]interface{}

	// Processor Alias for Process func
	Processor func(session Session, message string)(string, error)
)

//changed
//func sampleProcessor(session Session, message string, w http.ResponseWriter)(string,error){
func sampleProcessor(session Session, message string)(string,error){
  //write an HTTP client to fetch the text from the internet.

  url := "https://maps.googleapis.com/maps/api/place/textsearch/json?query="+message+"%20point%20of%20interest&language=en&key=AIzaSyAx8hZhRJBtys7xOB1-j4QP3JhON034v14"

    // Create an http client
    // Specify a timeout to limit the requests made by this client.
  	placesClient := http.Client{
  		Timeout: time.Second * 100, // Maximum of 2 secs
  	}

    // return a request suitable for use by the client
  	req, err := http.NewRequest(http.MethodGet, url, nil)
  	if err != nil {
  		log.Fatal(err)
  	}

    // A User-Agent has been set in the HTTP request's header. It lets remote servers understand what kind of traffic it is receiving.
  	req.Header.Set("User-Agent", "attractions")

    // Send an http request and wait for an http response
  	res, getErr := placesClient.Do(req)
  	if getErr != nil {
  		log.Fatal(getErr)
  	}

    // read from response
  	body, readErr := ioutil.ReadAll(res.Body)
  	if readErr != nil {
  		log.Fatal(readErr)
  	}

    // Decode the json data, unmarshal the data into the pointer to places
  	jsonErr := json.Unmarshal(body, &places)
  	if jsonErr != nil {
  		log.Fatal(jsonErr)
  	}

    // If the output array is empty. An error has occurred.
     if len(places.Results)==0{
        return "", fmt.Errorf("Not a valid city/country")
     }else{

    final := ""
    for _, result := range places.Results {
      s := strconv.FormatFloat(result.Rating, 'f', -1, 32)
      // final += "\n"
      final +=  "{<br> Attraction: " + result.Name +  "<br> Address: " + result.Formatted + "<br> Rating: " + s + "<br>  } <br> <br>"
    }
    //
    // json.NewEncoder(w).Encode(people1.Results)
    //json.NewEncoder(w).Encode(final)
    return final, nil}

  //
  //
  //
	// // Make sure a history key is defined in the session which points to a slice of strings
	// _, historyFound := session["history"]
	// if !historyFound {
	// 	session["history"] = []string{}
	// }
  //
	// // Fetch the history from session and cast it to an array of strings
	// history, _ := session["history"].([]string)
  //
	// // Make sure the message is unique in history
	// for _, m := range history {
	// 	if strings.EqualFold(m, message) {
	// 		return "", fmt.Errorf("You've already ordered %s before!", message)
	// 	}
	// }
  //
	// // Add the message in the parsed body to the messages in the session
	// history = append(history, message)
  //
	// // Form a sentence out of the history in the form Message 1, Message 2, and Message 3
	// l := len(history)
	// wordsForSentence := make([]string, l)
	// copy(wordsForSentence, history)
	// if l > 1 {
	// 	wordsForSentence[l-1] = "and " + wordsForSentence[l-1]
	// }
	// sentence := strings.Join(wordsForSentence, ", ")
  //
	// // Save the updated history to the session
	// session["history"] = history

	//return fmt.Sprintf(people1), nil
}




// withLog Wraps HandlerFuncs to log requests to Stdout
func withLog(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c := httptest.NewRecorder()
		fn(c, r)
		log.Printf("[%d] %-4s %s\n", c.Code, r.Method, r.URL.Path)

		for k, v := range c.HeaderMap {
			w.Header()[k] = v
		}
		w.WriteHeader(c.Code)
		c.Body.WriteTo(w)
	}
}

// writeJSON Writes the JSON equivilant for data into ResponseWriter w
func writeJSON(w http.ResponseWriter, data JSON) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// ProcessFunc Sets the processor of the chatbot
func ProcessFunc(p Processor) {
	processor = p
}

// handleWelcome Handles /welcome and responds with a welcome message and a generated UUID
func handleWelcome(w http.ResponseWriter, r *http.Request) {
	// Generate a UUID.
	hasher := md5.New()
	hasher.Write([]byte(strconv.FormatInt(time.Now().Unix(), 10)))
	uuid := hex.EncodeToString(hasher.Sum(nil))

	// Create a session for this UUID
	sessions[uuid] = Session{}

	// Write a JSON containg the welcome message and the generated UUID
	writeJSON(w, JSON{
		"uuid":    uuid,
		"message": WelcomeMessage,
	})
}

func handleChat(w http.ResponseWriter, r *http.Request) {
	// Make sure only POST requests are handled
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST requests are allowed.", http.StatusMethodNotAllowed)
		return
	}

	// Make sure a UUID exists in the Authorization header
	uuid := r.Header.Get("Authorization")
	if uuid == "" {
		http.Error(w, "Missing or empty Authorization header.", http.StatusUnauthorized)
		return
	}

	// Make sure a session exists for the extracted UUID
	session, sessionFound := sessions[uuid]
	if !sessionFound {
		http.Error(w, fmt.Sprintf("No session found for: %v.", uuid), http.StatusUnauthorized)
		return
	}

	// Parse the JSON string in the body of the request
	data := JSON{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, fmt.Sprintf("Couldn't decode JSON: %v.", err), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Make sure a message key is defined in the body of the request
	_, messageFound := data["message"]
	if !messageFound {
		http.Error(w, "Missing message key in body.", http.StatusBadRequest)
		return
	}

	// Process the received message
	message, err := processor(session, data["message"].(string))

  // Specify the error and the status code
  if err!=nil{
    http.Error(w,err.Error(),422)
    return
  }


	//Write a JSON containg the processed response
	writeJSON(w, JSON{
		"message": message,
	})
}

// handle Handles /
func handle(w http.ResponseWriter, r *http.Request) {
	body :=
		"<!DOCTYPE html><html><head><title>Chatbot</title></head><body><pre style=\"font-family: monospace;\">\n" +
			"Available Routes:\n\n" +
			"  GET  /welcome -> handleWelcome\n" +
			"  POST /chat    -> handleChat\n" +
			"  GET  /        -> handle        (current)\n" +
			"</pre></body></html>"
	w.Header().Add("Content-Type", "text/html")
	fmt.Fprintln(w, body)
}

// Engage Gives control to the chatbot
func Engage(addr string) error {
	// HandleFuncs
	mux := http.NewServeMux()
	mux.HandleFunc("/welcome", withLog(handleWelcome))
	mux.HandleFunc("/chat", withLog(handleChat))
	mux.HandleFunc("/", withLog(handle))

	// Start the server
	return http.ListenAndServe(addr, cors.CORS(mux))
}
func main() {

  // url := "https://maps.googleapis.com/maps/api/place/textsearch/json?query=egypt%20point%20of%20interest&language=en&key=AIzaSyAx8hZhRJBtys7xOB1-j4QP3JhON034v14"
  //
  // 	spaceClient := http.Client{
  // 		Timeout: time.Second * 100, // Maximum of 2 secs
  // 	}
  //
  // 	req, err := http.NewRequest(http.MethodGet, url, nil)
  // 	if err != nil {
  // 		log.Fatal(err)
  // 	}
  //
  // 	req.Header.Set("User-Agent", "spacecount-tutorial")
  //
  // 	res, getErr := spaceClient.Do(req)
  // 	if getErr != nil {
  // 		log.Fatal(getErr)
  // 	}
  //
  // 	body, readErr := ioutil.ReadAll(res.Body)
  // 	if readErr != nil {
  // 		log.Fatal(readErr)
  // 	}
  //
  // 	jsonErr := json.Unmarshal(body, &people1)
  // 	if jsonErr != nil {
  // 		log.Fatal(jsonErr)
  // 	}
  //
  //


	//Uncomment the following lines to customize the chatbot
	WelcomeMessage = "Hi, Where do you want to go?"
	//chatbot.ProcessFunc(chatbotProcess)

	// Use the PORT environment variable
	port := os.Getenv("PORT")
	// Default to 3000 if no PORT environment variable was defined
	if port == "" {
		port = "3000"
	}






	// Start the server
	fmt.Printf("Listening on port %s...\n", port)
	log.Fatalln(Engage(":" + port))
}
