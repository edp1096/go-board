// web/static/js/pages/admin-referrer-stats.js
document.addEventListener('DOMContentLoaded', function () {
    // Set up the form submit event
    const filterForm = document.getElementById('filter-form');
    if (filterForm) {
        filterForm.addEventListener('submit', function (e) {
            e.preventDefault();
            const days = document.getElementById('days').value;
            const limit = document.getElementById('limit').value;
            window.location.href = `/admin/referrer-stats?days=${days}&limit=${limit}`;
        });
    }

    // Load Chart.js if not already loaded
    if (typeof Chart === 'undefined') {
        loadScript('https://cdn.jsdelivr.net/npm/chart.js@3.9.1/dist/chart.min.js', renderCharts);
    } else {
        renderCharts();
    }
});

// Function to dynamically load a script
function loadScript(url, callback) {
    const script = document.createElement('script');
    script.type = 'text/javascript';
    script.src = url;
    script.onload = callback;
    document.head.appendChild(script);
}

// Function to render all charts
function renderCharts() {
    renderDailyChart();
}

// Function to render the daily visits chart
function renderDailyChart() {
    const chartContainer = document.getElementById('daily-chart');
    if (!chartContainer) return;

    // Get time stats data from the template
    const timeStatsElement = document.getElementById('time-stats-data');
    let timeStatsData = [];

    if (timeStatsElement) {
        try {
            timeStatsData = JSON.parse(timeStatsElement.textContent);
        } catch (e) {
            console.error('Error parsing time stats data:', e);
        }
    } else {
        // If the element doesn't exist, try to get data via AJAX
        fetchTimeStatsData().then(data => {
            if (data && data.timeStats) {
                renderDailyChartWithData(data.timeStats);
            }
        });
        return;
    }

    renderDailyChartWithData(timeStatsData);
}

// Function to render the daily chart with the provided data
function renderDailyChartWithData(timeStatsData) {
    const chartContainer = document.getElementById('daily-chart');
    if (!chartContainer || !timeStatsData) return;

    if (timeStatsData.length === 0) {
        chartContainer.innerHTML = '<p class="text-center text-gray-500 mt-8">데이터가 없습니다</p>';
        return;
    }

    // Prepare chart data
    const labels = timeStatsData.map(item => item.date);
    const counts = timeStatsData.map(item => item.count);

    // Create chart
    const ctx = document.createElement('canvas');
    chartContainer.innerHTML = '';
    chartContainer.appendChild(ctx);

    new Chart(ctx, {
        type: 'line',
        data: {
            labels: labels,
            datasets: [{
                label: '방문 수',
                data: counts,
                backgroundColor: 'rgba(59, 130, 246, 0.2)',
                borderColor: 'rgba(59, 130, 246, 1)',
                borderWidth: 2,
                pointBackgroundColor: 'rgba(59, 130, 246, 1)',
                tension: 0.1
            }]
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            scales: {
                y: {
                    beginAtZero: true,
                    ticks: {
                        precision: 0
                    }
                }
            },
            plugins: {
                tooltip: {
                    callbacks: {
                        title: function (tooltipItems) {
                            return tooltipItems[0].label;
                        },
                        label: function (context) {
                            return `방문 수: ${context.parsed.y}`;
                        }
                    }
                }
            }
        }
    });
}

// Function to fetch time stats data via AJAX
function fetchTimeStatsData() {
    const days = document.getElementById('days')?.value || 30;
    const limit = document.getElementById('limit')?.value || 10;

    return fetch(`/api/admin/referrer-stats?days=${days}&limit=${limit}`)
        .then(response => response.json())
        .then(data => {
            if (data.success) {
                return data;
            } else {
                console.error('Error fetching referrer stats:', data.message);
                return null;
            }
        })
        .catch(error => {
            console.error('Error fetching referrer stats:', error);
            return null;
        });
}