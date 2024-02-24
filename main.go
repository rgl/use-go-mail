package main

import (
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/wneessen/go-mail"
)

var indexTemplate = template.Must(template.New("Index").Parse(`<!DOCTYPE html>
<html>
<head>
<title>use-go-mail</title>
<style>
body {
	font-family: monospace;
	color: #555;
	background: #e6edf4;
	padding: 1.25rem;
	margin: 0;
}
label {
	display: inline-block;
	margin-bottom: 0.5em;
	min-width: 6em;
}
input[type="text"], input[type="password"] {
	min-width: 18em;
}
.error {
	color: red;
}
</style>
</head>
<body>
	<form method="post">
		<div>
			<label for="address">Address:</label>
			<input type="text" id="address" name="address" value="{{.Address}}" required>
		</div>
		<div>
			<label for="username">Username:</label>
			<input type="text" id="username" name="username" value="{{.Username}}" required>
		</div>
		<div>
			<label for="password">Password:</label>
			<input type="password" id="password" name="password" value="{{.Password}}" required>
		</div>
		<div>
			<label for="from">From:</label>
			<input type="text" id="from" name="from" value="{{.From}}" required>
		</div>
		<div>
			<label for="to">To:</label>
			<input type="text" id="to" name="to"  value="{{.To}}" required>
		</div>
		<div>
			<input type="submit" value="Send Mail">
			{{if .Error}}<span class="error">{{.Error}}</span>{{end}}
			{{if .Status}}<span>{{.Status}}</span>{{end}}
		</div>
	</form>
</body>
</html>
`))

type indexData struct {
	Address  string
	Username string
	Password string
	From     string
	To       string
	Error    string
	Status   string
}

func parseAddress(addr string) (tls bool, hostname string, port int, err error) {
	u, err := url.Parse(addr)
	if err != nil {
		return false, "", 0, fmt.Errorf("invalid address: %w", err)
	}
	if u.Scheme != "smtp" && u.Scheme != "smtps" {
		return false, "", 0, fmt.Errorf("invalid address scheme: %v", u.Scheme)
	}
	tls = u.Scheme != "smtp"
	hostname = u.Hostname()
	if u.Port() == "" {
		if tls {
			port = mail.DefaultPortSSL
		} else {
			port = mail.DefaultPort
		}
	} else {
		port, err = strconv.Atoi(u.Port())
		if err != nil {
			return false, "", 0, fmt.Errorf("invalid port: %w", err)
		}
	}
	return
}

func sendMail(addr string, username string, password string, from string, to string, subject string, body string) (string, error) {
	tls, hostname, port, err := parseAddress(addr)
	if err != nil {
		return "", err
	}
	options := []mail.Option{
		mail.WithUsername(username),
		mail.WithPassword(password),
	}
	if tls {
		options = append(options,
			mail.WithTLSPortPolicy(mail.TLSMandatory),
			mail.WithSSLPort(false),
			mail.WithPort(port),
		)
	} else {
		options = append(options,
			mail.WithTLSPortPolicy(mail.NoTLS),
			mail.WithPort(port),
		)
	}
	client, err := mail.NewClient(hostname, options...)
	if err != nil {
		return "", err
	}
	m := mail.NewMsg()
	if err := m.From(from); err != nil {
		return "", fmt.Errorf("failed to set FROM address: %w", err)
	}
	if err := m.To(to); err != nil {
		return "", fmt.Errorf("failed to set TO address: %w", err)
	}
	m.Subject(subject)
	m.SetBodyString(mail.TypeTextPlain, body)
	if err := client.DialAndSend(m); err != nil {
		return "", fmt.Errorf("failed to send mail: %w", err)
	}
	return fmt.Sprintf("Successfully sent mail: %s", subject), nil
}

func stringOrDefault(value string, defaultValue string) string {
	if value != "" {
		return value
	}
	return defaultValue
}

func main() {
	log.SetFlags(0)

	var listenAddress = flag.String("listen", ":8000", "Listen address.")

	flag.Parse()

	if flag.NArg() != 0 {
		flag.Usage()
		log.Fatalf("\nERROR You MUST NOT pass any positional arguments")
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("%s %s%s\n", r.Method, r.Host, r.URL)

		if r.URL.Path != "/" {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}

		data := indexData{
			Address:  os.Getenv("MAIL_SERVER_ADDR"),
			Username: "use-go-mail@localhost",
			Password: "password",
			From:     "use-go-mail@localhost",
			To:       "use-go-mail@localhost",
		}

		if r.Method == http.MethodPost {
			data.Address = stringOrDefault(r.PostFormValue("address"), data.Address)
			data.Username = stringOrDefault(r.PostFormValue("username"), data.Username)
			data.Password = stringOrDefault(r.PostFormValue("password"), data.Password)
			data.From = stringOrDefault(r.PostFormValue("from"), data.From)
			data.To = stringOrDefault(r.PostFormValue("to"), data.To)
			status, err := sendMail(
				data.Address,
				data.Username,
				data.Password,
				data.From,
				data.To,
				fmt.Sprintf("use-go-mail %s", time.Now().Format("2006-01-02T15:04:05.000000Z07:00")),
				"Hello, World!")
			if err != nil {
				data.Error = err.Error()
			}
			data.Status = status
		}

		w.Header().Set("Content-Type", "text/html")

		err := indexTemplate.ExecuteTemplate(w, "Index", data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	fmt.Printf("Listening at http://%s\n", *listenAddress)

	err := http.ListenAndServe(*listenAddress, nil)
	if err != nil {
		log.Fatalf("Failed to ListenAndServe: %v", err)
	}
}
