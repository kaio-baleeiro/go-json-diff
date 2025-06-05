package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	jsonpatch "github.com/evanphx/json-patch/v5"
	"github.com/wI2L/jsondiff"
)

// ProductDetails representa a estrutura aninhada "details".
// Usamos `omitempty` para que campos com valor zero (para números, bools) ou vazios (para slices/strings)
// não sejam incluídos no JSON final se não forem explicitamente definidos ou se resultarem em zero/vazio.
// Isso é útil para manter o JSON limpo, especialmente para campos que podem ser removidos ou adicionados.
type ProductDetails struct {
	DPI      int  `json:"dpi,omitempty"`
	Buttons  int  `json:"buttons,omitempty"`  // Estava em json1, mas não em json2 (será removido do resultado)
	Wireless bool `json:"wireless,omitempty"` // Não estava em json1, mas sim em json2 (será adicionado)
}

// Product representa a estrutura principal do JSON do produto.
type Product struct {
	Name    string         `json:"name"`
	Price   float64        `json:"price"`
	Active  bool           `json:"active"` // Note: se 'active' fosse opcional e false fosse um valor válido vs ausente, um ponteiro *bool seria melhor.
	Details ProductDetails `json:"details"`
	Tags    []string       `json:"tags"`
	Stock   int            `json:"stock,omitempty"` // Não estava em json1, mas sim em json2 (será adicionado)
}

// loadProductFromFile carrega um arquivo JSON para a struct Product.
// Retorna a struct populada e os bytes originais do arquivo (úteis para aplicar o patch).
func loadProductFromFile(filePath string) (*Product, []byte, error) {
	jsonDataBytes, err := os.ReadFile(filePath) // Go >= 1.16: os.ReadFile
	if err != nil {
		return nil, nil, fmt.Errorf("erro ao ler arquivo %s: %w", filePath, err)
	}

	var product Product
	if err := json.Unmarshal(jsonDataBytes, &product); err != nil {
		return nil, jsonDataBytes, fmt.Errorf("erro ao fazer unmarshal do JSON de %s: %w", filePath, err)
	}
	return &product, jsonDataBytes, nil
}

func main() {
	// --- 1. Carregar os JSONs para as structs e obter os bytes originais ---
	product1, json1OriginalBytes, err := loadProductFromFile("json1.json")
	if err != nil {
		log.Fatalf("Falha ao carregar json1.json: %v", err)
	}

	product2, _, err := loadProductFromFile("json2.json") // Bytes originais de json2 não são estritamente necessários aqui,
	// pois a comparação será entre product1 e product2.
	if err != nil {
		log.Fatalf("Falha ao carregar json2.json: %v", err)
	}

	fmt.Println("--- Product1 (carregado da struct): ---")
	p1Json, _ := json.MarshalIndent(product1, "", "  ")
	fmt.Println(string(p1Json))

	fmt.Println("\n--- Product2 (carregado da struct): ---")
	p2Json, _ := json.MarshalIndent(product2, "", "  ")
	fmt.Println(string(p2Json))

	// --- 2. Gerar o Patch (RFC 6902) comparando as structs ---
	// A biblioteca jsondiff.Compare pode lidar diretamente com as structs.
	patch, err := jsondiff.Compare(product1, product2)
	if err != nil {
		log.Fatalf("Erro ao gerar o patch JSON comparando structs: %v", err)
	}

	if len(patch) == 0 {
		fmt.Println("\nOs produtos (baseados nas structs) são idênticos. Nenhum patch gerado.")
		// Opcionalmente, salvar product1 (ou json1OriginalBytes) se for o caso.
		// Por exemplo, para garantir que o resultado esteja formatado pela struct:
		// finalResultBytes, _ := json.MarshalIndent(product1, "", "  ")
		// ioutil.WriteFile("json_resultante.json", finalResultBytes, 0644)
		return
	}

	patchBytes, _ := json.MarshalIndent(patch, "", "  ")
	fmt.Println("\n--- JSON Patch (RFC 6902) Gerado a partir das Structs: ---")
	fmt.Println(string(patchBytes))

	// --- 3. Aplicar o Patch aos bytes originais do json1 ---
	// É importante aplicar o patch aos bytes originais do json1,
	// pois o patch pode referenciar caminhos que existem no JSON original
	// mas que poderiam ser omitidos na struct (ou ter valor zero que `omitempty` esconderia
	// se tivéssemos re-serializado product1 antes de aplicar o patch).
	patchObj, err := jsonpatch.DecodePatch(patchBytes)
	if err != nil {
		log.Fatalf("Erro ao decodificar o patch JSON: %v", err)
	}
	resultAfterPatchBytes, err := patchObj.Apply(json1OriginalBytes)
	if err != nil {
		log.Fatalf("Erro ao aplicar o patch JSON aos bytes originais: %v", err)
	}

	// --- 4. Carregar o resultado do patch para a struct Product ---
	// Isso garante que o resultado final esteja em conformidade com o seu modelo de dados
	// e "limita" os campos aos definidos na struct Product.
	var finalProduct Product
	if err := json.Unmarshal(resultAfterPatchBytes, &finalProduct); err != nil {
		log.Fatalf("Erro ao fazer unmarshal do resultado do patch para a struct Product: %v", err)
	}

	// --- 5. Salvar ou usar o JSON resultante (a partir da struct finalProduct) ---
	finalOutputBytes, err := json.MarshalIndent(finalProduct, "", "  ")
	if err != nil {
		log.Fatalf("Erro ao formatar o JSON resultante da struct finalProduct: %v", err)
	}

	err = os.WriteFile("json_resultante.json", finalOutputBytes, 0644) // Go >= 1.16: os.WriteFile
	if err != nil {
		log.Fatalf("Erro ao salvar json_resultante.json: %v", err)
	}

	fmt.Println("\n--- JSON Resultante (a partir da struct final, salvo em json_resultante.json): ---")
	fmt.Println(string(finalOutputBytes))

	// Para modificar o arquivo1 original:
	// err = ioutil.WriteFile("json1.json", finalOutputBytes, 0644)
	// if err != nil {
	//    log.Fatalf("Erro ao sobrescrever json1.json: %v", err)
	// }
	// fmt.Println("\njson1.json foi atualizado com as diferenças, em conformidade com a struct Product.")
}
