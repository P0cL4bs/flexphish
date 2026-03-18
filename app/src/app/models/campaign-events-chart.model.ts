import {
    ApexChart,
    ApexNonAxisChartSeries,
    ApexXAxis,
    ApexStroke,
    ApexLegend,
    ApexTooltip,
    ApexDataLabels
} from 'ng-apexcharts';

export type ChartOptions = {
    series: ApexNonAxisChartSeries;
    chart: ApexChart;
    xaxis: ApexXAxis;
    stroke: ApexStroke;
    legend: ApexLegend;
    tooltip: ApexTooltip;
    dataLabels: ApexDataLabels;
};