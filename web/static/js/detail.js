let currentAssetData = null;
let selectedConfig = null;
let chart = null;
let series = null;
let supportLines = [];

document.addEventListener('DOMContentLoaded', async () => {
    const symbol = window.location.pathname.split('/').pop();
    document.getElementById('asset-title').innerText = `Analizando ${symbol.toUpperCase()}...`;

    try {
        const res = await fetch(`/api/v1/assets/${symbol}/analysis`);
        if (!res.ok) {
            const errData = await res.json().catch(() => ({}));
            throw new Error(errData.error || `HTTP ${res.status}: Fallo al conectar con la API de análisis`);
        }
        
        currentAssetData = await res.json();
        document.getElementById('asset-title').innerText = `${symbol.toUpperCase()} - $${currentAssetData.current_price.toFixed(4)}`;
        
        // Ensure data is sorted ascending by time and no duplicates exist (Required by TradingView)
        currentAssetData.candles.sort((a, b) => a.time - b.time);
        const uniqueCandles = [];
        for (let i = 0; i < currentAssetData.candles.length; i++) {
            if (i === 0 || currentAssetData.candles[i].time !== currentAssetData.candles[i-1].time) {
                uniqueCandles.push(currentAssetData.candles[i]);
            }
        }
        currentAssetData.candles = uniqueCandles;

        renderScore(currentAssetData.score);
        renderConfigs(currentAssetData.suggested_configs);
        initChart(currentAssetData.candles, currentAssetData.supports, currentAssetData.resistances);
        
    } catch (e) {
        console.error(e);
        document.getElementById('asset-title').innerText = `Error: ${e.message}`;
    }
});

function renderScore(scoreData) {
    const box = document.getElementById('score-box');
    const val = document.getElementById('score-val');
    const trend = document.getElementById('score-trend');

    val.innerText = `${scoreData.TotalScore}/100`;
    trend.innerText = `Tendencia: ${scoreData.Trend}`;

    if (scoreData.TotalScore >= 60) {
        box.style.borderLeftColor = 'var(--accent-green)';
        val.style.color = 'var(--accent-green)';
    } else if (scoreData.TotalScore >= 40) {
        box.style.borderLeftColor = '#fbbf24'; // yellow
        val.style.color = '#fbbf24';
    } else {
        box.style.borderLeftColor = 'var(--accent-red)';
        val.style.color = 'var(--accent-red)';
    }
}

function renderConfigs(configs) {
    const container = document.getElementById('configs-container');
    container.innerHTML = '';

    if (!configs || configs.length === 0) {
        container.innerHTML = '<div class="text-muted text-sm">No se pudieron calcular configuraciones aptas.</div>';
        return;
    }

    configs.forEach((cfg, idx) => {
        const card = document.createElement('div');
        card.className = 'config-card';
        if (idx === 0) {
            card.classList.add('selected');
            selectedConfig = cfg;
        }

        card.innerHTML = `
            <div style="font-weight: 600; margin-bottom: 8px;">${cfg.type}</div>
            <div style="font-size: 13px; color: var(--text-muted); display: grid; grid-template-columns: 1fr 1fr; gap: 4px;">
                <div>Rango: <span style="color:white">$${cfg.lower_price.toFixed(4)} - $${cfg.upper_price.toFixed(4)}</span></div>
                <div>Grids: <span style="color:white">${cfg.num_grids}</span></div>
                <div>Spacing: <span style="color:white">${cfg.grid_spacing_percent.toFixed(2)}%</span></div>
                <div style="grid-column: span 2;">Est. ROI (365d): <span style="color:var(--accent-green)">${cfg.estimated_roi.toFixed(2)}%</span></div>
            </div>
        `;

        card.onclick = () => {
            document.querySelectorAll('.config-card').forEach(c => c.classList.remove('selected'));
            card.classList.add('selected');
            selectedConfig = cfg;
            drawGridOnChart(cfg);
        };

        container.appendChild(card);
    });
}

function initChart(candles, supports, resistances) {
    const container = document.getElementById('tvchart');
    chart = LightweightCharts.createChart(container, {
        layout: {
            background: { type: 'solid', color: 'transparent' },
            textColor: '#d1d5db',
        },
        grid: {
            vertLines: { color: 'rgba(255, 255, 255, 0.05)' },
            horzLines: { color: 'rgba(255, 255, 255, 0.05)' },
        },
        crosshair: { mode: LightweightCharts.CrosshairMode.Normal },
        rightPriceScale: { borderColor: 'rgba(255, 255, 255, 0.1)' },
        timeScale: { borderColor: 'rgba(255, 255, 255, 0.1)' },
    });

    series = chart.addCandlestickSeries({
        upColor: '#10b981',
        downColor: '#ef4444',
        borderVisible: false,
        wickUpColor: '#10b981',
        wickDownColor: '#ef4444',
    });

    series.setData(candles);

    // Draw S/R
    if (supports) {
        supports.forEach(lvl => {
            series.createPriceLine({ price: lvl.price, color: 'rgba(16, 185, 129, 0.4)', lineWidth: 1, lineStyle: 2, title: 'Sup' });
        });
    }
    if (resistances) {
        resistances.forEach(lvl => {
            series.createPriceLine({ price: lvl.price, color: 'rgba(239, 68, 68, 0.4)', lineWidth: 1, lineStyle: 2, title: 'Res' });
        });
    }

    if (currentAssetData.suggested_configs.length > 0) {
        drawGridOnChart(currentAssetData.suggested_configs[0]);
    }
}

function drawGridOnChart(cfg) {
    // Clear old lines
    supportLines.forEach(l => series.removePriceLine(l));
    supportLines = [];

    const step = (cfg.upper_price - cfg.lower_price) / cfg.num_grids;
    for (let i = 0; i <= cfg.num_grids; i++) {
        const p = cfg.lower_price + (step * i);
        const l = series.createPriceLine({
            price: p,
            color: '#3b82f6', // Azul brillante sólido
            lineWidth: 2,     // Más grueso
            lineStyle: 2,     // Dashed (Punteado)
            title: `Grid ${i}` // Etiqueta visible a la derecha
        });
        supportLines.push(l);
    }
}

async function runSimulation() {
    if (!selectedConfig) return;

    const capital = parseFloat(document.getElementById('sim-capital').value);
    const fee = parseFloat(document.getElementById('sim-fee').value) / 100.0;
    const symbol = window.location.pathname.split('/').pop();

    try {
        const res = await fetch(`/api/v1/assets/${symbol}/simulate`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ config: selectedConfig, capital: capital, fee: fee })
        });
        if (!res.ok) throw new Error('Simulation failed');
        
        const data = await res.json();
        
        document.getElementById('sim-results').style.display = 'block';
        document.getElementById('sim-trades').innerText = data.total_trades;
        document.getElementById('sim-comms').innerText = `$${data.total_commissions.toFixed(2)}`;
        document.getElementById('sim-net').innerText = `$${data.net_profit.toFixed(2)}`;
        document.getElementById('sim-roi').innerText = `${data.total_roi_percent.toFixed(2)}%`;
        document.getElementById('sim-dd').innerText = `${data.max_drawdown_percent.toFixed(2)}%`;

    } catch (e) {
        console.error(e);
        alert('Error ejecutando simulación.');
    }
}
