# GridBot — BingX Grid Spot Analyzer

**GridBot** es una herramienta avanzada de análisis financiero diseñada para identificar las mejores oportunidades de trading en el mercado Spot de **BingX** utilizando estrategias de **Grid Bot**.

## Características Principales

- **Universo de Activos**: Filtra automáticamente las 100 principales criptomonedas (excluyendo stablecoins) y verifica su disponibilidad en BingX Spot.
- **Motor Técnico**: Cálculo interno de indicadores clave:
  - **ATR**: Volatilidad relativa para dimensionar el grid.
  - **ADX**: Detección de mercados laterales (rango) vs tendencias.
  - **RSI**: Identificación de momentum neutral.
  - **Bollinger Bands**: Compresión de precio y posición relativa.
  - **Soportes y Resistencias**: Identificación automática de niveles estructurales.
  - **Choppiness & ER**: Medición de ruido y eficiencia del mercado.
- **Scoring Paramétrico**: El usuario elige qué indicadores influyen en la puntuación final.
- **Simulación Histórica**: Backtesting rápido de 7, 15 y 30 días para evaluar el rendimiento esperado.
- **Interfaz Premium**: Dashboard oscuro moderno con visualizaciones gráficas interactivas.

## Instalación

1. Asegúrate de tener instalado **Go 1.21+**.
2. Clona el repositorio y entra en la carpeta:
   ```bash
   cd GridBot
   ```
3. Instala las dependencias:
   ```bash
   go mod tidy
   ```
4. Configura las variables de entorno en el archivo `.env` (puedes usar el `.env.example` como base).
5. Compila y ejecuta:
   ```bash
   go run main.go
   ```
6. Abre tu navegador en `http://localhost:8080`.

## Cómo Funciona el Scoring

El sistema asigna una puntuación de 0 a 100 a cada activo basándose en los indicadores seleccionados:

- **ADX < 20**: Puntuación máxima (Mercado ideal en rango).
- **RSI 40-60**: Puntuación máxima (Neutralidad).
- **ATR 3-15%**: Puntuación máxima (Volatilidad saludable).
- **CHOP > 61.8**: Puntuación máxima (Mercado "picado" ideal para grillas).

Solo los activos con un **Score ≥ 50** son recomendados para operar.

## Configuración del Grid Sugerida

Para los activos recomendados, el sistema calcula:
- **Rango**: Basado en soportes y resistencias locales con un buffer de seguridad.
- **Grillas**: Optimizadas para que la rentabilidad neta por grilla cubra las comisiones de BingX (0.2% round-trip) y deje un beneficio neto saludable (>0.4%).

---
*Descargo de responsabilidad: Esta herramienta es solo para fines informativos y de análisis. El trading de criptomonedas conlleva riesgos significativos. No constituye asesoramiento financiero.*
