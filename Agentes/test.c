#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include "quiche.h"

#define CERT "(base 64 cert)"
#define KEY "(base 64 private key)"
#define SERVER_CERT "(base 64 cert)"

char* quic_get(const char *url) {
    // Certificado y clave privada
    const uint8_t *cert = (const uint8_t *) CERT;
    const size_t cert_len = strlen(CERT);
    const uint8_t *key = (const uint8_t *) KEY;
    const size_t key_len = strlen(KEY);

    // Certificado del servidor
    const uint8_t *server_cert = (const uint8_t *) SERVER_CERT;
    const size_t server_cert_len = strlen(SERVER_CERT);

    // Inicializar la biblioteca quiche
    quiche_get();

    // Establecer la configuración del cliente
    quiche_config *config = quiche_config_new(QUICHE_PROTOCOL_VERSION);

    // Establecer la configuración de la versión
    quiche_config_set_application_protos(config,
                                         (uint8_t *) QUICHE_H3_APPLICATION_PROTOCOL,
                                         sizeof(QUICHE_H3_APPLICATION_PROTOCOL) - 1);
    
    // Establecer el certificado y la clave privada desde la memoria
    quiche_config_load_cert_chain_from_pem_mem(config, cert, cert_len, key, key_len);

    // Establecer el certificado del servidor desde la memoria
    quiche_config_load_verify_locations_from_pem_mem(config, server_cert, server_cert_len);

    // Crear una nueva conexión
    quiche_conn *conn = quiche_connect(url, config);

    // Crear un nuevo flujo
    
    
}




void quic_post(const char *url, const char *data, size_t data_len) {
    // Implementa aquí la función para realizar una solicitud POST usando quiche
    printf("POST %s\n", url);
}

int main(int argc, char *argv[]) {
    if (argc < 2) {
        fprintf(stderr, "Uso: %s URL\n", argv[0]);
        return 1;
    }

    const char *url = argv[1];

    quic_get(url);

    const char *post_data = "datos para enviar en la solicitud POST";
    quic_post(url, post_data, strlen(post_data));

    return 0;
}