document.addEventListener("DOMContentLoaded", function() {
  
  // ==========================================
  // 1. RISK DISTRIBUTION (DONUT CHART)
  // ==========================================
  const donutEl = document.querySelector("#riskDonutChart");
  
  // Pastikan elemennya ada di halaman sebelum di-render
  if (donutEl) {
    // Tarik data dari HTML (dataset)
    const valExtreme = parseInt(donutEl.dataset.extreme) || 0;
    const valHigh = parseInt(donutEl.dataset.high) || 0;
    const valMedium = parseInt(donutEl.dataset.medium) || 0;
    const valLow = parseInt(donutEl.dataset.low) || 0;

    const riskOptions = {
      series: [valExtreme, valHigh, valMedium, valLow],
      labels: ['Extreme', 'High', 'Medium', 'Low'],
      chart: {
        type: 'donut',
        height: 320,
        fontFamily: 'Inter, sans-serif'
      },
      colors: ['#212529', '#ff3e1d', '#ffab00', '#71dd37'], 
      plotOptions: {
        pie: {
          donut: {
            size: '75%',
            labels: {
              show: true,
              name: { fontSize: '14px', color: '#6c757d' },
              value: { fontSize: '24px', fontWeight: 'bold', color: '#32475c' },
              total: {
                show: true,
                label: 'Total Assets',
                formatter: function (w) {
                  return w.globals.seriesTotals.reduce((a, b) => a + b, 0)
                }
              }
            }
          }
        }
      },
      dataLabels: { enabled: false },
      stroke: { width: 5, colors: ['#fff'] },
      legend: { position: 'bottom' }
    };
    new ApexCharts(donutEl, riskOptions).render();
  }

  // ==========================================
  // 2. INSPECTION FORECAST (BAR CHART)
  // ==========================================
  const barEl = document.querySelector("#inspectionBarChart");
  
  if (barEl) {
    // Tarik string JSON dari HTML dan jadikan array JavaScript
    const chartYears = JSON.parse(barEl.dataset.years || '[]');
    const chartCounts = JSON.parse(barEl.dataset.counts || '[]');

    const barOptions = {
      series: [{
        name: 'Inspections Due',
        data: chartCounts
      }],
      chart: {
        type: 'bar',
        height: 320,
        fontFamily: 'Inter, sans-serif',
        toolbar: { show: false }
      },
      plotOptions: {
        bar: {
          borderRadius: 6,
          columnWidth: '40%',
          distributed: true 
        }
      },
      colors: ['#696cff'], 
      dataLabels: { enabled: false },
      xaxis: {
        categories: chartYears,
        labels: { style: { colors: '#6c757d', fontSize: '13px' } }
      },
      yaxis: {
        labels: { style: { colors: '#6c757d', fontSize: '13px' } },
        title: { text: 'Equipment Count', style: { color: '#6c757d' } }
      },
      grid: { borderColor: '#e9ecef', strokeDashArray: 4 }
    };
    new ApexCharts(barEl, barOptions).render();
  }

});