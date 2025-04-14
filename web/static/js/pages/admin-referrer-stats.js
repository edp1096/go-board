// web/static/js/pages/admin-referrer-stats.js
document.addEventListener('DOMContentLoaded', function () {
    // 일별 방문 차트 초기화
    initDailyChart();

    // 레퍼러 타입 차트 초기화
    initReferrerTypeChart();

    // 필터 폼 이벤트 리스너
    document.getElementById('filter-form').addEventListener('submit', function (e) {
        e.preventDefault();
        const days = document.getElementById('days').value;
        const limit = document.getElementById('limit').value;
        const view = document.getElementById('view').value;

        window.location.href = `/admin/referrer-stats?days=${days}&limit=${limit}&view=${view}`;
    });
});

// 일별 방문 차트 초기화
function initDailyChart() {
    const timeStats = JSON.parse(document.getElementById('time-stats-data').value || '[]');
    if (timeStats.length === 0) return;

    // Chart.js 설정
    const dailyChartEl = document.getElementById('daily-chart');
    if (!dailyChartEl) return;

    const ctx = dailyChartEl.getContext('2d');
    new Chart(ctx, {
        type: 'line',
        data: {
            labels: timeStats.map(item => item.date),
            datasets: [{
                label: '방문 수',
                data: timeStats.map(item => item.count),
                borderColor: 'rgba(59, 130, 246, 1)', // Blue
                backgroundColor: 'rgba(59, 130, 246, 0.1)',
                borderWidth: 2,
                tension: 0.3,
                fill: true
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
                        label: function (context) {
                            return `방문 수: ${context.raw}`;
                        }
                    }
                }
            }
        }
    });
}

// 레퍼러 타입 차트 초기화
function initReferrerTypeChart() {
    const typeStats = JSON.parse(document.getElementById('type-stats-data').value || '[]');
    if (typeStats.length === 0) return;

    const typeChartEl = document.getElementById('type-chart');
    if (!typeChartEl) return;

    // 데이터 매핑 및 색상 설정
    const labels = typeStats.map(item => {
        switch (item.type) {
            case 'direct': return '직접 방문';
            case 'search': return '검색엔진';
            case 'social': return '소셜미디어';
            default: return '기타';
        }
    });

    const data = typeStats.map(item => item.count);

    const colors = typeStats.map(item => {
        switch (item.type) {
            case 'direct': return 'rgba(107, 114, 128, 0.8)'; // Gray
            case 'search': return 'rgba(16, 185, 129, 0.8)';  // Green
            case 'social': return 'rgba(59, 130, 246, 0.8)';  // Blue
            default: return 'rgba(245, 158, 11, 0.8)';        // Yellow
        }
    });

    // 차트 생성
    const ctx = typeChartEl.getContext('2d');
    new Chart(ctx, {
        type: 'doughnut',
        data: {
            labels: labels,
            datasets: [{
                data: data,
                backgroundColor: colors,
                borderWidth: 1
            }]
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            plugins: {
                legend: {
                    position: 'right'
                },
                tooltip: {
                    callbacks: {
                        label: function (context) {
                            const value = context.raw;
                            const percent = typeStats[context.dataIndex].percentTotal.toFixed(1);
                            return `${value} (${percent}%)`;
                        }
                    }
                }
            }
        }
    });
}