// go-java-manager.go
// Utilitário para manipular a versão do Java (instalar, trocar) via Go.
// Suporta Windows e MacOS.

package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// Detecta o sistema operacional
func getOS() string { return runtime.GOOS }

// Instala uma versão específica do Java
func installJava(version string) error {
	osType := getOS()
	var cmd *exec.Cmd
	log.Printf("Iniciando instalação do Java %s...", version)
	if osType == "windows" {
		cmd = exec.Command("choco", "install", fmt.Sprintf("temurin%s", version), "-y")
	} else if osType == "darwin" {
		cmd = exec.Command("brew", "install", fmt.Sprintf("openjdk@%s", version))
	} else {
		return fmt.Errorf("Sistema operacional não suportado: %s", osType)
	}
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	return cmd.Run()
}

// Troca a versão ativa do Java
func switchJava(version string) error {
	osType := getOS()
	var cmd *exec.Cmd
	log.Printf("Trocando JAVA_HOME para Java %s...", version)
	if osType == "windows" {
		// Busca o diretório real do Temurin instalado
		javaHome := ""
		var latestVersion string
		dirs, err := os.ReadDir("C:/Program Files/Eclipse Adoptium/")
		if err == nil {
			for _, d := range dirs {
				if d.IsDir() && (latestVersion == "" || d.Name() > latestVersion) && strings.Contains(d.Name(), version) {
					latestVersion = d.Name()
				}
			}
			if latestVersion != "" {
				javaHome = "C:/Program Files/Eclipse Adoptium/" + latestVersion
			}
		}
		if javaHome == "" {
			javaHome = fmt.Sprintf("C:/Program Files/Java/jdk-%s", version)
		}
		os.Setenv("JAVA_HOME", javaHome)
		// Tenta atualizar JAVA_HOME na sessão atual do PowerShell
		if pwsh := os.Getenv("PSModulePath"); pwsh != "" {
			fmt.Printf("$env:JAVA_HOME = '%s'\n", javaHome)
			fmt.Println("Copie e cole o comando acima no seu terminal para atualizar JAVA_HOME na sessão atual.")
		}
		// Define JAVA_HOME permanentemente para o usuário e para a máquina
		cmd = exec.Command("powershell", "-Command", fmt.Sprintf("[System.Environment]::SetEnvironmentVariable(\"JAVA_HOME\",\"%s\",\"User\"); [System.Environment]::SetEnvironmentVariable(\"JAVA_HOME\",\"%s\",\"Machine\")", javaHome, javaHome))
		if err := cmd.Run(); err != nil {
			log.Printf("Falha ao definir JAVA_HOME permanentemente: %v", err)
		}
		log.Printf("JAVA_HOME configurado para: %s", javaHome)
		log.Printf("ATENÇÃO: Para que o JAVA_HOME seja reconhecido na sessão atual do terminal, copie e cole o comando abaixo no seu PowerShell:")
		fmt.Printf("$env:JAVA_HOME = '%s'\n", javaHome)
		return nil
	} else if osType == "darwin" {
		// Para sessão atual
		macCmd := fmt.Sprintf("export JAVA_HOME=\"$(/usr/libexec/java_home -v %s)\"", version)
		os.Setenv("JAVA_HOME", fmt.Sprintf("$(/usr/libexec/java_home -v %s)", version))
		// Para futuras sessões, adiciona ao ~/.zshrc e ~/.bash_profile
		cmd = exec.Command("/bin/bash", "-c", fmt.Sprintf(
			"echo '%s' >> ~/.zshrc; echo '%s' >> ~/.bash_profile", macCmd, macCmd))
	} else {
		return fmt.Errorf("Sistema operacional não suportado: %s", osType)
	}
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	return cmd.Run()
}

// Executa um comando Java usando o JAVA_HOME configurado
func RunJavaCommand(args ...string) error {
	osType := getOS()
	javaBin := "java"
	if osType == "windows" {
		// Busca o diretório real do Temurin instalado
		javaHome := ""
		dirs, err := os.ReadDir("C:/Program Files/Eclipse Adoptium/")
		if err == nil {
			for _, d := range dirs {
				if d.IsDir() && (len(javaHome) == 0 || d.Name() > javaHome) {
					javaHome = "C:/Program Files/Eclipse Adoptium/" + d.Name()
				}
			}
		}
		if javaHome == "" {
			javaHome = os.Getenv("JAVA_HOME")
		}
		if javaHome != "" {
			javaBin = javaHome + "\\bin\\java.exe"
		}
	} else if osType == "darwin" {
		javaHome := os.Getenv("JAVA_HOME")
		if javaHome != "" {
			javaBin = javaHome + "/bin/java"
		}
	}
	cmd := exec.Command(javaBin, args...)
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	return cmd.Run()
}

// Verifica se está rodando como administrador no Windows
func isElevated() bool {
	if getOS() != "windows" {
		return true // Não se aplica
	}
	return exec.Command("fltmc").Run() == nil
}

// Verifica se uma ferramenta está instalada e atualizada
func checkAndUpdateTool(tool string, updateCmd []string) error {
	log.Printf("Verificando se '%s' está instalado...", tool)
	_, err := exec.LookPath(tool)
	if err != nil {
		return fmt.Errorf("Ferramenta '%s' não encontrada. Instale antes de continuar.", tool)
	}
	log.Printf("Atualizando '%s'...", tool)
	cmd := exec.Command(updateCmd[0], updateCmd[1:]...)
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Falha ao atualizar %s: %v", tool, err)
	}
	log.Printf("'%s' atualizado com sucesso.", tool)
	return nil
}

const javaVersion = "17"

func main() {
	osType := getOS()
	log.Printf("Sistema operacional detectado: %s", osType)
	if osType == "windows" && !isElevated() {
		log.Fatal("Execute este programa como Administrador para instalar Java via Chocolatey!")
	}
	var tools = map[string][]string{}
	if osType == "windows" {
		tools["choco"] = []string{"choco", "upgrade", "chocolatey", "-y", "--allow-downgrade", "--allow-global-confirmation"}
	} else if osType == "darwin" {
		tools["brew"] = []string{"brew", "update"}
	}
	for tool, updateCmd := range tools {
		if err := checkAndUpdateTool(tool, updateCmd); err != nil {
			log.Fatal(err)
		}
	}
	if err := installJava(javaVersion); err != nil {
		log.Printf("Falha na instalação do Java: %v", err)
	} else {
		log.Printf("Java instalado com sucesso!")
	}
	if err := switchJava(javaVersion); err != nil {
		log.Printf("Falha ao trocar versão do Java: %v", err)
	} else {
		log.Printf("Versão do Java trocada com sucesso!")
	}
	log.Printf("Executando 'java -version' com o JAVA_HOME configurado...")
	if err := RunJavaCommand("-version"); err != nil {
		log.Printf("Falha ao executar 'java -version': %v", err)
	}
}
