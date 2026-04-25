document.addEventListener('DOMContentLoaded', () => {
    const universeBtn = document.getElementById('refresh-universe');
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

    universeBtn.addEventListener('click', loadUniverse);

    async function loadUniverse() {
        universeBtn.disabled = true;
        universeBtn.textContent = 'Cargando...';
        try {
            const resp = await fetch('/api/assets/universe');
            allAssets = await resp.json();
            updateRankingTable([]); // Clear
            showNotification('Universo de activos cargado: ' + allAssets.length, 'success');
        } catch (err) {
            showNotification('Error al cargar activos', 'error');
        } finally {
            universeBtn.disabled = false;
            universeBtn.textContent = 'Cargar Activos';
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
        progressFill.style.width = '20%';
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

        // Simulation Stats
        const simStats = document.getElementById('sim-stats');
        simStats.innerHTML = `
            <div class="stat-box">
                <span class="stat-label">ROI Neto (${sim.period}d)</span>
                <span class="stat-value" style="color: ${sim.roi_net >= 0 ? 'var(--primary)' : 'var(--danger)'}">${sim.roi_net.toFixed(2)}%</span>
            </div>
            <div class="stat-box">
                <span class="stat-label">Capital Final ($1000)</span>
                <span class="stat-value">$${sim.final_capital.toFixed(2)}</span>
            </div>
            <div class="stat-box">
                <span class="stat-label">Operaciones</span>
                <span class="stat-value">${sim.estimated_operations}</span>
            </div>
            <div class="stat-box">
                <span class="stat-label">vs Buy & Hold</span>
                <span class="stat-value">${sim.buy_and_hold_roi.toFixed(2)}%</span>
            </div>
        `;

        renderChart(sim.history, gc.lower_range, gc.upper_range);

        modal.classList.remove('hidden');
    }

    function renderChart(history, lower, upper) {
        const ctx = document.getElementById('price-chart').getContext('2d');
        if (priceChart) priceChart.destroy();

        const labels = history.map(h => new Date(h.timestamp).toLocaleDateString());
        const prices = history.map(h => h.close);

        priceChart = new Chart(ctx, {
            type: 'line',
            data: {
                labels: labels,
                datasets: [
                    {
                        label: 'Precio',
                        data: prices,
                        borderColor: '#6B7FFF',
                        borderWidth: 2,
                        pointRadius: 0,
                        fill: false
                    },
                    {
                        label: 'Grid Upper',
                        data: Array(labels.length).fill(upper),
                        borderColor: 'rgba(0, 212, 170, 0.5)',
                        borderDash: [5, 5],
                        pointRadius: 0,
                        fill: false
                    },
                    {
                        label: 'Grid Lower',
                        data: Array(labels.length).fill(lower),
                        borderColor: 'rgba(231, 76, 60, 0.5)',
                        borderDash: [5, 5],
                        pointRadius: 0,
                        fill: false
                    }
                ]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                scales: {
                    y: { grid: { color: '#252D3D' }, ticks: { color: '#8892A4' } },
                    x: { grid: { display: false }, ticks: { color: '#8892A4' } }
                },
                plugins: {
                    legend: { display: false }
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
