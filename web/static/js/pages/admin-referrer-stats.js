// WHOIS 모달 관련 자바스크립트 함수들

// showIPDetails 함수 구현 - IP 클릭 시 상세 정보 모달 표시
// notuse. not_use
function showIPDetails(ip, userAgents) {
    const modal = document.getElementById('whois-modal');
    const modalTitle = document.getElementById('whois-modal-title');
    const loading = document.getElementById('whois-loading');
    const error = document.getElementById('whois-error');
    const errorMessage = document.getElementById('whois-error-message');
    const content = document.getElementById('whois-content');
    const summary = document.getElementById('whois-summary');
    const whoisRawData = document.getElementById('whois-raw-data');

    // 모달 제목 설정 및 표시
    modalTitle.textContent = `IP 주소 상세 정보: ${ip}`;
    modal.classList.add('show');
    modal.style.display = 'flex';

    // User-Agent 정보 표시 
    loading.style.display = 'none';
    content.style.display = 'block';

    // WHOIS API 호출 먼저
    fetch(`/api/whois?ip=${ip}`)
        .then(response => response.json())
        .then(data => {
            if (data.success) {
                displayWhoisInfo(data.data, 'ip', summary);
            } else {
                summary.innerHTML = `<div class="py-2">IP 정보를 가져올 수 없습니다.</div>`;
            }

            // User-Agent 정보 표시
            let uaHTML = `
                <div class="mt-4">
                    <h4 class="text-sm font-medium mb-2">User-Agent 정보</h4>
                    <div class="bg-accent rounded-lg p-3 overflow-y-auto max-h-64">
            `;

            try {
                // 문자열이면 파싱, 이미 배열이면 그대로 사용
                const agents = typeof userAgents === 'string'
                    ? JSON.parse(userAgents)
                    : userAgents;

                if (agents && agents.length) {
                    agents.forEach(ua => {
                        // 간단한 UA 분석
                        const isBot = /bot|crawl|spider|slurp|search/i.test(ua);
                        const isMobile = /mobile|android|iphone|ipad/i.test(ua);

                        let icon = 'fa-globe';
                        if (isBot) icon = 'fa-robot';
                        else if (isMobile) icon = 'fa-mobile-alt';
                        else if (/chrome/i.test(ua)) icon = 'fa-chrome';
                        else if (/firefox/i.test(ua)) icon = 'fa-firefox-browser';
                        else if (/safari/i.test(ua)) icon = 'fa-safari';

                        uaHTML += `
                            <div class="mb-2 p-2 border border-gray-200 rounded">
                                <p><i class="fas ${icon} mr-2"></i>${ua}</p>
                            </div>
                        `;
                    });
                } else {
                    uaHTML += `<p>User-Agent 정보가 없습니다.</p>`;
                }
            } catch (e) {
                uaHTML += `<p>User-Agent 정보를 읽을 수 없습니다.</p>`;
                console.error('User-Agent 파싱 오류:', e);
            }

            uaHTML += `</div></div>`;
            whoisRawData.innerHTML = uaHTML;
        })
        .catch(err => {
            summary.innerHTML = `<div class="py-2">IP 정보를 가져올 수 없습니다.</div>`;
            console.error('IP 정보 조회 오류:', err);

            // User-Agent만 표시
            whoisRawData.innerHTML = `<div class="mt-2">User-Agent 정보를 불러올 수 없습니다.</div>`;
        });
}

// 봇/사람 비율 차트 초기화
function initVisitorTypeChart() {
    // 선택된 레퍼러에서 봇/사람 수 집계
    const botCount = topReferrers.reduce((sum, ref) => sum + (ref.uaStats?.botCount || 0), 0);
    const humanCount = topReferrers.reduce((sum, ref) => sum + (ref.uaStats?.humanCount || 0), 0);

    const ctx = document.getElementById('visitor-type-chart').getContext('2d');
    new Chart(ctx, {
        type: 'pie',
        data: {
            labels: ['봇', '사람'],
            datasets: [{
                data: [botCount, humanCount],
                backgroundColor: [
                    'rgba(54, 162, 235, 0.8)',
                    'rgba(75, 192, 192, 0.8)'
                ],
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
                            const total = botCount + humanCount;
                            const percent = Math.round((value / total) * 100);
                            return `${context.label}: ${value} (${percent}%)`;
                        }
                    }
                }
            }
        }
    });
}

// 브라우저 분포 차트 초기화
function initBrowserChart() {
    // 브라우저 데이터 집계
    const browsers = {};

    topReferrers.forEach(ref => {
        if (ref.uaStats?.browsers) {
            Object.entries(ref.uaStats.browsers).forEach(([browser, count]) => {
                browsers[browser] = (browsers[browser] || 0) + count;
            });
        }
    });

    const labels = Object.keys(browsers);
    const data = Object.values(browsers);

    // 색상 맵
    const browserColors = {
        'Chrome': 'rgba(66, 133, 244, 0.8)', // Google Blue
        'Firefox': 'rgba(255, 89, 0, 0.8)',  // Firefox Orange
        'Safari': 'rgba(0, 122, 255, 0.8)',  // Safari Blue
        'Edge': 'rgba(0, 120, 215, 0.8)',    // Edge Blue
        'IE': 'rgba(0, 120, 215, 0.8)',      // IE Blue
        'Bot': 'rgba(128, 128, 128, 0.8)',   // Gray for bots
        'Unknown': 'rgba(169, 169, 169, 0.8)' // Dark Gray for unknown
    };

    // 색상 배열 생성
    const colors = labels.map(browser =>
        browserColors[browser] || 'rgba(83, 166, 131, 0.8)' // Default Green
    );

    const ctx = document.getElementById('browser-chart').getContext('2d');
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
                            const total = data.reduce((sum, val) => sum + val, 0);
                            const percent = Math.round((value / total) * 100);
                            return `${context.label}: ${value} (${percent}%)`;
                        }
                    }
                }
            }
        }
    });
}

// 모바일/PC 비율 차트 초기화
function initDeviceChart() {
    // 디바이스 유형 집계
    const mobileCount = topReferrers.reduce((sum, ref) => sum + (ref.uaStats?.mobileCount || 0), 0);
    const desktopCount = topReferrers.reduce((sum, ref) => sum + (ref.uaStats?.desktopCount || 0), 0);

    const ctx = document.getElementById('device-chart').getContext('2d');
    new Chart(ctx, {
        type: 'pie',
        data: {
            labels: ['모바일', '데스크톱'],
            datasets: [{
                data: [mobileCount, desktopCount],
                backgroundColor: [
                    'rgba(255, 159, 64, 0.8)', // Orange for Mobile
                    'rgba(54, 162, 235, 0.8)'  // Blue for Desktop
                ],
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
                            const total = mobileCount + desktopCount;
                            const percent = Math.round((value / total) * 100);
                            return `${context.label}: ${value} (${percent}%)`;
                        }
                    }
                }
            }
        }
    });
}

function showWhoisInfo(value, type, ua) {
    const modal = document.getElementById('whois-modal');
    const modalTitle = document.getElementById('whois-modal-title');
    const loading = document.getElementById('whois-loading');
    const error = document.getElementById('whois-error');
    const errorMessage = document.getElementById('whois-error-message');
    const content = document.getElementById('whois-content');
    const summary = document.getElementById('whois-summary');
    const whoisRawData = document.getElementById('whois-raw-data');
    const uaRawData = document.getElementById('ua-raw-data');

    // 모달 표시 (내용은 비워둔 상태로)
    modalTitle.textContent = type === 'ip' ? `IP 주소 정보: ${value}` : `도메인 정보: ${value}`;

    // 다른 요소들은 모달이 표시된 후에 상태 변경
    modal.classList.add('show');
    modal.style.display = 'flex';

    // 모달이 표시된 후 내용 초기화 (깜빡임 방지)
    setTimeout(() => {
        // 모달 초기화
        loading.style.display = 'flex';
        error.style.display = 'none';
        content.style.display = 'none';

        // WHOIS API 호출
        fetch(`/api/whois?${type}=${value}`)
            .then(response => {
                if (!response.ok) {
                    throw new Error('WHOIS 정보를 가져오는데 실패했습니다');
                }
                return response.json();
            })
            .then(data => {
                if (!data.success) {
                    throw new Error(data.message || '알 수 없는 오류가 발생했습니다');
                }

                // 로딩 숨기기
                loading.style.display = 'none';

                // WHOIS 정보 표시
                displayWhoisInfo(data.data, type, summary);

                // JSON 데이터 포맷팅 (가독성 향상)
                let formattedData;
                if (data.data.rawData) {
                    // JSON 형식인지 확인 시도
                    try {
                        const jsonData = JSON.parse(data.data.rawData);
                        formattedData = formatJsonDisplay(jsonData);
                    } catch (e) {
                        // JSON이 아니면 원본 그대로 표시
                        formattedData = data.data.rawData;
                    }
                } else {
                    formattedData = '원시 데이터가 없습니다';
                }

                whoisRawData.innerHTML = formattedData;
                uaRawData.innerHTML = ua.toString();

                // 내용 표시
                content.style.display = 'block';
            })
            .catch(err => {
                // 로딩 숨기기
                loading.style.display = 'none';

                // 오류 메시지 표시
                errorMessage.textContent = err.message;
                error.style.display = 'block';

                console.error('WHOIS 정보 조회 오류:', err);
            });
    }, 50); // 아주 짧은 지연 (모달이 표시된 후 콘텐츠 업데이트)
}

// JSON 데이터를 가독성 좋게 포맷팅하는 함수
function formatJsonDisplay(jsonData) {
    // 주요 정보를 추출하여 표 형식으로 표시
    const tableData = extractImportantInfo(jsonData);

    // 접을 수 있는 원본 JSON 데이터 섹션 생성
    const collapsibleSection = `
        <div class="mt-4">
            <details class="rounded p-2">
                <summary class="cursor-pointer font-medium text-sm py-1">원본 JSON 데이터 보기</summary>
                <pre class="text-xs overflow-auto mt-2 p-2 rounded max-h-64">${JSON.stringify(jsonData, null, 2)}</pre>
            </details>
        </div>
    `;

    return tableData + collapsibleSection;
}

// JSON 데이터에서 중요 정보 추출 및 표 형식으로 변환
function extractImportantInfo(jsonData) {
    let result = `
<div class="overflow-x-auto">
    <table class="min-w-full text-sm">`;

    result += `
<thead>
    <tr>
        <th class="text-left py-2 px-3 w-1/3">속성</th><th class="text-left py-2 px-3">
            값
        </th>
    </tr>
</thead>
<tbody>
`;

    // IP 정보용 주요 필드 추출
    if (jsonData.ipVersion) {
        result += addTableRow('IP 버전', jsonData.ipVersion);
        result += addTableRow('IP 범위', `${jsonData.startAddress} - ${jsonData.endAddress}`);
        result += addTableRow('네트워크', jsonData.handle || jsonData.name || '-');
        result += addTableRow('상태', Array.isArray(jsonData.status) ? jsonData.status.join(', ') : (jsonData.status || '-'));
        result += addTableRow('국가', jsonData.country || '-');

        // 조직 정보
        if (jsonData.entities && jsonData.entities.length > 0) {
            for (const entity of jsonData.entities) {
                if (entity.vcardArray && entity.vcardArray.length > 1) {
                    const vcard = entity.vcardArray[1];
                    for (const prop of vcard) {
                        if (prop[0] === 'fn') {
                            result += addTableRow('조직', prop[3] || '-');
                            break;
                        }
                    }
                }
            }
        }
    }
    // 도메인 정보용 필드 추출
    else if (jsonData.domain) {
        result += addTableRow('도메인', jsonData.domain);
        result += addTableRow('등록기관', jsonData.registrar || '-');
        result += addTableRow('생성일', jsonData.createdDate || '-');
        result += addTableRow('만료일', jsonData.expiryDate || '-');
        result += addTableRow('네임서버', Array.isArray(jsonData.nameServers) ? jsonData.nameServers.join('<br>') : (jsonData.nameServers || '-'));
        result += addTableRow('상태', Array.isArray(jsonData.status) ? jsonData.status.join('<br>') : (jsonData.status || '-'));
    }

    result += '</tbody></table></div>';
    return result;
}

// 테이블 행 생성 헬퍼 함수
function addTableRow(label, value) {
    return `<tr class="border-t border-gray-200">
        <td class="py-2 px-3 align-top font-medium">${label}</td>
        <td class="py-2 px-3 align-top">${value}</td>
    </tr>`;
}

// WHOIS 정보 표시 함수
function displayWhoisInfo(data, type, container) {
    container.innerHTML = '';

    if (type === 'ip') {
        // IP 주소 정보 표시
        const ipInfoHTML = `
            <div class="grid grid-cols-1 md:grid-cols-2 gap-3">
                <div>
                    <p class="text-xs">IP 주소</p>
                    <p class="font-medium">${data.ipAddress || data.query || '-'}</p>
                </div>
                <div>
                    <p class="text-xs">위치</p>
                    <p class="font-medium">${data.country || '-'}${data.city ? ', ' + data.city : ''}</p>
                </div>
                <div>
                    <p class="text-xs">네트워크</p>
                    <p class="font-medium">${data.network || data.asn || '-'}</p>
                </div>
                <div>
                    <p class="text-xs">조직</p>
                    <p class="font-medium">${data.organization || data.isp || '-'}</p>
                </div>
            </div>
        `;
        container.innerHTML = ipInfoHTML;
    } else {
        // 도메인 정보 표시
        const domainInfoHTML = `
            <div class="grid grid-cols-1 md:grid-cols-2 gap-3">
                <div>
                    <p class="text-xs">도메인</p>
                    <p class="font-medium">${data.domain || '-'}</p>
                </div>
                <div>
                    <p class="text-xs">등록기관</p>
                    <p class="font-medium">${data.registrar || '-'}</p>
                </div>
                <div>
                    <p class="text-xs">생성일</p>
                    <p class="font-medium">${data.createdDate || '-'}</p>
                </div>
                <div>
                    <p class="text-xs">만료일</p>
                    <p class="font-medium">${data.expiryDate || '-'}</p>
                </div>
                <div class="md:col-span-2">
                    <p class="text-xs">네임서버</p>
                    <p class="font-medium">${Array.isArray(data.nameServers) ? data.nameServers.join(', ') : (data.nameServers || '-')}</p>
                </div>
                <div class="md:col-span-2">
                    <p class="text-xs">상태</p>
                    <p class="font-medium">${Array.isArray(data.status) ? data.status.join(', ') : (data.status || '-')}</p>
                </div>
            </div>
        `;
        container.innerHTML = domainInfoHTML;
    }
}

// 모달 닫기 함수
function closeWhoisModal() {
    const modal = document.getElementById('whois-modal');

    // 페이드 아웃 효과를 위한 트랜지션 클래스 추가
    modal.classList.add('closing');

    // 트랜지션이 완료된 후 실제로 모달 숨기기
    setTimeout(() => {
        modal.classList.remove('show', 'closing');
        modal.style.display = 'none';
    }, 300); // 트랜지션 시간과 일치
}


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


// web/static/js/pages/admin-referrer-stats.js
document.addEventListener('DOMContentLoaded', function () {
    // ESC 키로 모달 닫기
    document.addEventListener('keydown', function (e) {
        if (e.key === 'Escape') {
            closeWhoisModal();
        }
    });

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
        const dnsCheckbox = document.getElementById('dns');
        const showDNS = dnsCheckbox && dnsCheckbox.checked ? 'true' : 'false';

        window.location.href = `/admin/referrer-stats?days=${days}&limit=${limit}&view=${view}&dns=${showDNS}`;
    });

    // 상위 레퍼러 데이터 가져오기 (전역 변수)
    window.topReferrers = [];

    // API로 데이터 가져오기
    fetch(`/api/admin/referrer-stats?mode=top&days=${document.getElementById('days').value}`)
        .then(response => response.json())
        .then(data => {
            if (data.success && data.topReferrers) {
                window.topReferrers = data.topReferrers;

                // 새로운 차트 초기화
                initVisitorTypeChart();
                initBrowserChart();
                initDeviceChart();
            }
        })
        .catch(err => console.error('레퍼러 데이터 로드 오류:', err));
});