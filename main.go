package main

import (
	"encoding/json"
	"fmt"
	"net/http" // cria o servidor HTTP e lida com as requisições

	_ "github.com/lib/pq"

	"database/sql"
	"log"
)

// estrutura do usuário que queremos salvar
type Usuarios struct {
	Id       int    `json:"id"`
	Nome     string `json:"nome"`
	Email    string `json:"email"`
	Telefone string `json:"telefone"`
}

// variável de conexão com o banco de dados
var db *sql.DB

func Create(w http.ResponseWriter, r *http.Request) {

	// verificar se a requisicao utiliza o metodo POST
	if r.Method != http.MethodPost {
		fmt.Println("Método não permitido")
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	// criar uma variavel vazia para armazenar os dados dos usuarios
	u := Usuarios{}

	// decodificar o JSON recebido no corpo da requisicao e armazenar os valores nos campos da struct Usuario
	var err error
	err = json.NewDecoder(r.Body).Decode(&u)
	if err != nil {
		fmt.Println("Erro ao decodificar JSON:", err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	// inserir os dados do usuario no banco de dados
	_, err = db.Exec(
		"INSERT INTO usuarios (nome, email, telefone) VALUES ($1, $2, $3)",
		u.Nome, u.Email, u.Telefone,
	)
	if err != nil {
		log.Println("Erro ao inserir dados no banco:", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	// informar que o cadastro foi realizado com sucesso
	fmt.Fprintln(w, "Cadastro realizado com sucesso")
	w.WriteHeader(http.StatusCreated) // 201
}

func Read(w http.ResponseWriter, r *http.Request) {
	// verificar se a requisição utiliza o método GET
	if r.Method != http.MethodGet {
		fmt.Print("Método não permitido")
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	// executar uma consulta para buscar todos os usuarios
	rows, err := db.Query(
		"SELECT id, nome, email, telefone FROM usuarios",
	)
	if err != nil {
		log.Println("Erro ao consultar o banco dados:", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// garantir que as linhas da consulta sejam fechadas ao final da função
	defer rows.Close()

	// criar uma slice vazia para armazenar os usuarios encontrados
	usuarios := []Usuarios{}

	// percorrer cada linha retornada pelo banco de dados
	for rows.Next() {
		u := Usuarios{}

		// copiar os valores da linha para os campos da struct Usuario
		err := rows.Scan(&u.Id, &u.Nome, &u.Email, &u.Telefone)

		if err != nil {
			log.Println("Erro ao ler dados do usuário:", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		// adicionar o usuario atual a slice de usuarios
		usuarios = append(usuarios, u)

	}

	err = rows.Err()
	if err != nil {
		log.Println("Erro ao ler linhas do banco de dados:", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// informar que a resposta sera um JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// converter a lista de usuarios para JSON e enviar na resposta
	err = json.NewEncoder(w).Encode(usuarios)
	if err != nil {
		log.Println("Erro ao codificar JSON:", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func Update(w http.ResponseWriter, r *http.Request) {
	u := Usuarios{}

	var err error
	err = json.NewDecoder(r.Body).Decode(&u)
	if err != nil {
		fmt.Println("Erro ao decodificar JSON:", err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	// verificar se utiliza o método PUT
	if r.Method != http.MethodPut {
		fmt.Println("Método não permitido.")
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	res, err := db.Exec(
		"UPDATE usuarios SET nome = $1, email = $2, telefone = $3 WHERE id = $4",
		&u.Nome, &u.Email, &u.Telefone, &u.Id,
	)
	if err != nil {
		log.Println("Erro ao atualizar o usuário:", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// verificar se alguma linha foi alterada no banco
	// garante que, caso o id não exista, não será considerado um sucesso na atualização
	linhasModificadas, err := res.RowsAffected()
	if err != nil {
		log.Println("Erro ao verificar linhas modificadas:", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	if linhasModificadas == 0 {
		http.Error(w, "Nenhum usuário encontrado para atualizar", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "Usuário atualizado com sucesso")
}

func Delete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		fmt.Println("Método não permitido")
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	// pegar o id da url
	id := r.PathValue("id")

	res, err := db.Exec("DELETE FROM usuarios WHERE id = $1", id)
	if err != nil {
		log.Println("Erro ao deletar o usuário:", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	linhasModificadas, err := res.RowsAffected()
	if err != nil {
		log.Println("Erro ao verificar linhas modificadas:", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	if linhasModificadas == 0 {
		http.Error(w, "Nenhum usuário encontrado para deletar", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "Usuário deletado com sucesso")

}

func main() {
	// string de conexão com o banco
	connString := "postgres://postgres:stone@localhost:5432/crudb?sslmode=disable"

	// abrir a conexao com o postgres utilizando o driver pq
	var err error // evita shadowing
	db, err = sql.Open("postgres", connString)
	if err != nil {
		log.Println("Erro ao abrir a conexão com o banco de dados:", err)
		return
	}

	// garantir que a conexão com o banco seja fechada quando o programa terminar
	defer db.Close()

	// verificar se o banco de dados esta acessivel
	err = db.Ping()
	if err != nil {
		log.Println("Erro ao conectar com o banco de dados:", err)
		return
	}

	fmt.Println("Conectado ao banco de dados com sucesso")

	// roteador HTTP responsável por analisar as requisições HTTP recebidas (método e URL)
	// e direcioná-las para a função correta (o handler)
	mux := http.NewServeMux()

	mux.HandleFunc("POST /cadastrar", Create)
	mux.HandleFunc("GET /usuarios", Read)
	mux.HandleFunc("PUT /usuarios", Update)
	mux.HandleFunc("DELETE /usuarios/{id}", Delete)

	// iniciar o servidor HTTP na porta 8080
	// (err não é obrigatorio) e verificar se o servidor iniciou corretamente
	err = http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Println("Erro ao iniciar o servidor HTTP:", err)
		return
	}
}
