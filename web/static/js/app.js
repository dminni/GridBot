document.addEventListener('DOMContentLoaded', () => {
    // const universeBtn = document.getElementById('refresh-universe');
    const analysisForm = document.getElementById('analysis-form');
    const rankingBody = document.getElementById('ranking-body');
    const discardedList = document.getElementById('discarded-list');
    const progressContainer = document.getElementById('progress-container');
    const progressFill = document.getElementById('progress-fill');
    const progressText = document.getElementById('progress-text');
    const searchInput = document.getElementById('search-input');
    const modal = document.getElementById('detail-modal');
    const closeModal = document.querySelector('.close-modal');

    let allAssets = [];
    let rankedAssets = [];
    let priceChart = null;

    // Load universe on start
    loadUniverse();

    // universeBtn.addEventListener('click', loadUniverse);

    async function loadUniverse() {
        try {
            const resp = await fetch('/api/assets/universe');
            allAssets = await resp.json();
            showNotification('Universo de activos cargado: ' + allAssets.length, 'success');
        } catch (err) {
            showNotification('Error al cargar activos', 'error');
        }
    }

    analysisForm.addEventListener('submit', async (e) => {
        e.preventDefault();
        const formData = new FormData(analysisForm);
        const indicators = formData.getAll('indicators');

        if (indicators.length === 0) {
            alert('Seleccioná al menos un indicador');
            return;
        }

        progressContainer.classList.remove('hidden');
        progressFill.style.width = '10%';
        progressText.textContent = 'Verificando activos...';

        if (allAssets.length === 0) {
            await loadUniverse();
        }

        progressFill.style.width = '30%';
        progressText.textContent = 'Calculando indicadores...';

        try {
            const resp = await fetch('/api/analysis/run', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ indicators })
            });
            
            progressFill.style.width = '80%';
            const result = await resp.json();
            
            if (resp.status !== 200) {
                alert(result.error);
                return;
            }

            rankedAssets = result.ranked_assets || [];
            updateRankingTable(rankedAssets);
            updateDiscardedList(result.discarded_assets || []);
            
            document.getElementById('last-update').textContent = 'Último análisis: ' + new timeFormatter().format(new Date());
            showNotification('Análisis completado', 'success');
        } catch (err) {
            showNotification('Error en el análisis', 'error');
        } finally {
            progressFill.style.width = '100%';
            setTimeout(() => progressContainer.classList.add('hidden'), 1000);
        }
    });

    function updateRankingTable(assets) {
        rankingBody.innerHTML = '';
        assets.sort((a, b) => b.analysis.score - a.analysis.score).forEach((a, i) => {
            const row = document.createElement('tr');
            row.innerHTML = `
                <td>${i + 1}</td>
                <td class="symbol-col"><strong>${a.symbol}</strong><br><small>${a.name}</small></td>
                <td>$${a.current_price.toLocaleString()}</td>
                <td><span class="score-badge ${getScoreClass(a.analysis.score)}">${a.analysis.score}/100</span></td>
                <td>${a.analysis.indicators.RSI?.toFixed(1) || '--'}</td>
                <td>${a.analysis.indicators.ATR?.toFixed(1) || '--'}%</td>
                <td>${a.analysis.indicators.ADX?.toFixed(1) || '--'}</td>
                <td>${a.analysis.indicators.CHOP?.toFixed(1) || '--'}</td>
                <td>${a.analysis.indicators.ER?.toFixed(2) || '--'}</td>
                <td>${a.analysis.status}</td>
                <td><button class="btn btn-secondary btn-sm view-detail" data-symbol="${a.symbol}">Ver Detalle</button></td>
            `;
            row.addEventListener('click', () => showDetail(a.symbol));
            rankingBody.appendChild(row);
        });
    }

    function updateDiscardedList(assets) {
        discardedList.innerHTML = '';
        assets.forEach(a => {
            const card = document.createElement('div');
            card.className = 'discarded-card';
            card.innerHTML = `<strong>${a.symbol}</strong> - Score: ${a.analysis?.score || 0}`;
            discardedList.appendChild(card);
        });
    }

    async function showDetail(symbol) {
        const resp = await fetch(`/api/asset/${symbol}`);
        const data = await resp.json();
        const asset = data.asset;
        const sim = data.simulation;

        document.getElementById('modal-title').textContent = asset.symbol + ' / USDT';
        document.getElementById('modal-price').textContent = '$' + asset.current_price.toLocaleString();
        document.getElementById('justification-text').textContent = asset.analysis.justification;

        // Stats
        const stats = document.getElementById('indicators-stats');
        stats.innerHTML = '';
        const important = ['ADX', 'RSI', 'ATR', 'HV', 'CHOP', 'ER'];
        important.forEach(key => {
            if (asset.analysis.indicators[key]) {
                const box = document.createElement('div');
                box.className = 'stat-box';
                box.innerHTML = `<span class="stat-label">${key}</span><span class="stat-value">${asset.analysis.indicators[key].toFixed(2)}</span>`;
                stats.appendChild(box);
            }
        });

        // Grid Config
        const gc = asset.analysis.grid_config;
        document.getElementById('grid-lower').textContent = '$' + gc.lower_range.toLocaleString();
        document.getElementById('grid-upper').textContent = '$' + gc.upper_range.toLocaleString();
        document.getElementById('grid-amp').textContent = gc.amplitude_pct.toFixed(2) + '%';
        document.getElementById('grid-count').textContent = gc.recommended_grids;
        document.getElementById('grid-profit').textContent = gc.profit_per_grid_net.toFixed(2) + '%';

        const altBody = document.getElementById('alt-configs-body');
        altBody.innerHTML = '';
        gc.alternative_configs.forEach(c => {
            altBody.innerHTML += `<tr><td>${c.type}</td><td>${c.grids}</td><td>${c.profit_per_grid.toFixed(2)}%</td></tr>`;
        });

        // Simulations
        const simulations = data.simulations;
        const periods = [7, 15, 30];
        
        simulations.forEach((sim, idx) => {
            const period = periods[idx];
            const statsContainer = document.getElementById(`sim-stats-${period}`);
            if (sim) {
                statsContainer.innerHTML = `
                    <span class="roi-badge" style="background: ${sim.roi_net >= 0 ? 'rgba(0,212,170,0.2)' : 'rgba(231,76,60,0.2)'}; color: ${sim.roi_net >= 0 ? 'var(--primary)' : 'var(--danger)'}">
                        ${sim.roi_net >= 0 ? '+' : ''}${sim.roi_net.toFixed(2)}%
                    </span>
                    <span class="ops-count">${sim.estimated_operations} ops</span>
                    <span class="ops-count">B&H: ${sim.buy_and_hold_roi.toFixed(2)}%</span>
                `;
                renderChart(`price-chart-${period}`, sim.history, gc, sim.trades || []);
            } else {
                statsContainer.innerHTML = '<span style="color:var(--text-sec)">Sin datos</span>';
            }
        });

        modal.classList.remove('hidden');
    }

    function renderChart(canvasId, history, gridConfig, trades) {
        const ctx = document.getElementById(canvasId).getContext('2d');
        
        if (window.charts && window.charts[canvasId]) {
            window.charts[canvasId].destroy();
        }
        if (!window.charts) window.charts = {};

        const labels = history.map(h => {
            const d = new Date(h.timestamp);
            return d.getDate() + '/' + (d.getMonth() + 1);
        });
        const prices = history.map(h => h.close);

        const lower = gridConfig.lower_range;
        const upper = gridConfig.upper_range;
        const grids = gridConfig.recommended_grids;
        const step = (upper - lower) / grids;

        // Build a timestamp-to-index map for placing trade markers
        const tsToIdx = {};
        history.forEach((h, i) => { tsToIdx[h.timestamp] = i; });

        const datasets = [
            {
                label: 'Precio',
                data: prices,
                borderColor: '#6B7FFF',
                borderWidth: 2,
                pointRadius: 0,
                fill: false,
                order: 1
            }
        ];

        // Upper/Lower range bands with shaded fill between them
        datasets.push({
            label: 'Upper',
            data: Array(labels.length).fill(upper),
            borderColor: '#00D4AA',
            borderWidth: 2,
            pointRadius: 0,
            fill: false,
            order: 2
        });
        datasets.push({
            label: 'Lower',
            data: Array(labels.length).fill(lower),
            borderColor: '#E74C3C',
            borderWidth: 2,
            pointRadius: 0,
            fill: '-1', // fill between Lower and Upper
            backgroundColor: 'rgba(107, 127, 255, 0.05)',
            order: 2
        });

        // Add shaded area for High and Low prices (wicks/volatility)
        datasets.push({
            label: 'High',
            data: history.map(h => h.high),
            borderColor: 'transparent',
            borderWidth: 0,
            pointRadius: 0,
            fill: false,
            order: 4
        });
        datasets.push({
            label: 'Low',
            data: history.map(h => h.low),
            borderColor: 'transparent',
            borderWidth: 0,
            pointRadius: 0,
            fill: '-1', // fill between Low and High
            backgroundColor: 'rgba(255, 255, 255, 0.05)',
            order: 4
        });

        // Grid lines (subtle)
        const maxGridLines = Math.min(grids - 1, 30); // cap to avoid too many lines
        const gridStep = Math.max(1, Math.floor((grids - 1) / maxGridLines));
        for (let i = gridStep; i < grids; i += gridStep) {
            datasets.push({
                data: Array(labels.length).fill(lower + i * step),
                borderColor: 'rgba(136, 146, 164, 0.15)',
                borderWidth: 1,
                borderDash: [2, 3],
                pointRadius: 0,
                fill: false,
                order: 3
            });
        }

        // Buy markers (green triangles)
        const buyData = Array(labels.length).fill(null);
        const sellData = Array(labels.length).fill(null);
        
        trades.forEach(t => {
            const idx = tsToIdx[t.timestamp];
            if (idx !== undefined) {
                if (t.type === 'buy') {
                    buyData[idx] = t.price;
                } else {
                    sellData[idx] = t.price;
                }
            }
        });

        datasets.push({
            label: 'Compra',
            data: buyData,
            borderColor: '#00D4AA',
            backgroundColor: '#00D4AA',
            pointStyle: 'triangle',
            pointRadius: 6,
            pointHoverRadius: 8,
            showLine: false,
            order: 0
        });

        datasets.push({
            label: 'Venta',
            data: sellData,
            borderColor: '#F5A623',
            backgroundColor: '#F5A623',
            pointStyle: 'rectRot',
            pointRadius: 6,
            pointHoverRadius: 8,
            showLine: false,
            order: 0
        });

        window.charts[canvasId] = new Chart(ctx, {
            type: 'line',
            data: { labels, datasets },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                scales: {
                    y: { 
                        grid: { color: 'rgba(37, 45, 61, 0.5)' }, 
                        ticks: { color: '#8892A4', font: { size: 10 } } 
                    },
                    x: { 
                        grid: { display: false }, 
                        ticks: { color: '#8892A4', font: { size: 10 }, maxRotation: 0 } 
                    }
                },
                plugins: {
                    legend: { 
                        display: true,
                        position: 'top',
                        labels: { 
                            color: '#8892A4', 
                            font: { size: 10 },
                            usePointStyle: true,
                            filter: (item) => ['Precio','Upper','Lower','Compra','Venta'].includes(item.text)
                        }
                    },
                    tooltip: { enabled: true }
                }
            }
        });
    }

    closeModal.addEventListener('click', () => modal.classList.add('hidden'));
    window.addEventListener('click', (e) => { if (e.target === modal) modal.classList.add('hidden'); });

    function getScoreClass(score) {
        if (score >= 80) return 'score-excellent';
        if (score >= 65) return 'score-good';
        if (score >= 50) return 'score-regular';
        return 'score-bad';
    }

    function showNotification(msg, type) {
        console.log(`[${type}] ${msg}`);
        // Implement toast here if needed
    }

    class timeFormatter {
        format(date) {
            return date.getHours().toString().padStart(2, '0') + ':' + date.getMinutes().toString().padStart(2, '0');
        }
    }

    searchInput.addEventListener('input', (e) => {
        const val = e.target.value.toUpperCase();
        const rows = rankingBody.querySelectorAll('tr');
        rows.forEach(row => {
            const sym = row.querySelector('.symbol-col strong').textContent;
            row.style.display = sym.includes(val) ? '' : 'none';
        });
    });
});
