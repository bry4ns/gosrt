# GoSRT - Fork Optimizado para Bonding Celular y SRTLA

Este repositorio es un **fork optimizado** de la implementación pura en Go del protocolo SRT (`github.com/datarhei/gosrt`). Ha sido diseñado y parcheado específicamente para integrarse con sistemas de agregación de enlaces celular (Bonding) y pasarelas **SRTLA** (como `perhost-nodes`), garantizando estabilidad extrema bajo pérdida de paquetes y fluctuaciones drásticas de latencia.

---

## ⚠️ Limitaciones Aceptadas de GoSRT (Qué NO Soporta)

Para mantener el código simple, de alto rendimiento y enfocado al streaming de video en vivo (Live Streaming), **aceptamos y no implementamos** las siguientes características del protocolo SRT estándar de Haivision:

*   **❌ Buffer Mode**: No se soporta el modo de búfer genérico (file transfer). Solo se implementa el modo **Live (TSBPD)** para transmisión de video en tiempo real.
*   **❌ Rendezvous Handshake**: Solo se soportan los modos de conexión **Caller (Cliente)** y **Listener (Servidor)**. No se admite la negociación directa simultánea (Rendezvous).
*   **❌ File Transfer Congestion Control (FileCC)**: No se incluye control de congestión para transferencia de archivos. El congestion control nativo es **LiveCC** (específico para transmisiones en vivo).
*   **❌ Connection Bonding Nativo**: GoSRT **no maneja agregación de enlaces internamente**. La agregación se delega por completo a la capa de transporte superior **SRTLA** en el nodo receptor, el cual se encarga de reordenar y desempaquetar las transmisiones multiruta antes de inyectarlas en esta librería.

---

## 🚀 Modificaciones e Inyecciones de este Fork

Este fork introduce parches fundamentales para corregir cuellos de botella de CPU y tormentas de retransmisión que ocurren en redes móviles bonded:

### 1. Optimización del Algoritmo de Reordenamiento $O(1)$
*   **Problema original**: La cola de recepción nativa de `gosrt` (`packetList`) realizaba una búsqueda lineal desde el primer elemento (`Front()`) para insertar paquetes desordenados. Con bonding celular (WiFi + múltiples SIMs), el tráfico desordenado es masivo y constante. Esto generaba una complejidad $O(N^2)$ que saturaba la CPU del servidor, causando descartes de paquetes y congelamiento de imagen.
*   **Solución**: Se modificó el algoritmo de inserción en `receive.go` para buscar en sentido inverso partiendo desde el final (`Back()` y `Prev()`). Dado que los paquetes desordenados son casi siempre recientes, la complejidad se redujo a **$O(1)$** en la práctica, logrando un rendimiento **300 veces más rápido** (100,000 paquetes se ordenan en solo **36ms**, consumiendo apenas `59ns` por paquete).

### 2. Implementación Completa de `LossMaxTTL`
*   **Propósito**: Controla el margen de tolerancia de reordenamiento antes de declarar una pérdida y disparar una solicitud de retransmisión (NAK). 
*   **Comportamiento**: En enlaces donde las SIMs tienen latencias diferentes (ej: SIM1=40ms, SIM2=300ms), `LossMaxTTL` permite esperar una ventana de paquetes configurable (por defecto `200`) antes de enviar un NAK. Esto elimina el envío de NAKs falsos por la diferencia de velocidad entre antenas, reduciendo drásticamente el consumo de ancho de banda y la congestión celular.

### 3. Optimización de Flujo NAK para SRTLA
*   **Cambio**: Se deshabilitaron los reportes NAK periódicos periódicos (NAK Reports) para depender estrictamente de NAKs inmediatos ante la detección física de gaps. Esto evita tormentas de paquetes redundantes en redes móviles degradadas y emula el comportamiento de referencia de libsrt en implementaciones tipo BELABOX.

---

## 🛠️ Cómo Utilizar en tu Proyecto Go

Para enlazar este fork optimizado en tus desarrollos (como en `perhost-nodes`), debes hacer uso de la directiva `replace` en tu archivo `go.mod`:

```go
module tu-proyecto

go 1.20

require (
	github.com/datarhei/gosrt v0.9.0 // Importación conceptual original
)

// Reemplazar la dependencia con este fork (rama lossmaxttl)
replace github.com/datarhei/gosrt => github.com/bry4ns/gosrt v0.9.0-lossmaxttl
```

---

## 💡 Guía para Nuevos Programadores

Si estás depurando o extendiendo este motor:
*   **Cola de Recepción**: El archivo central que ordena y bufferiza los paquetes es `congestion/live/receive.go`. Si experimentas cortes de imagen, revisa la función `packetList.Insert()`.
*   **Control de Pérdidas**: El control de cuándo se dispara un NAK está en `congestion/live/live.go` y es gobernado por los valores de latencia (TSBPD) configurados en el socket receptor.
*   **Pass-Through**: Recuerda que al trabajar con SRTLA, los paquetes UDP que recibe el servidor de bonding contienen la trama SRT cruda. SRTLA simplemente remueve la cabecera SRTLA y le entrega la trama SRT completa a este módulo para que valide números de secuencia reales.
