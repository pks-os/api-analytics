<script lang="ts">
	import { onMount } from 'svelte';
	import { ColumnIndex } from '../../../lib/consts';

	function getDevice(userAgent: string): string {
		if (userAgent === null) {
			return 'Unknown';
		} else if (userAgent.match(/iPhone/)) {
			return 'iPhone';
		} else if (userAgent.match(/Android/)) {
			return 'Android';
		} else if (userAgent.match(/Tizen\//)) {
			return 'Samsung';
		} else if (userAgent.match(/Macintosh/)) {
			return 'Mac';
		} else if (userAgent.match(/Windows/)) {
			return 'PC';
		} else {
			return 'Other';
		}
	}

	function devicePlotLayout() {
		return {
			title: false,
			autosize: true,
			margin: { r: 35, l: 70, t: 10, b: 20, pad: 0 },
			hovermode: 'closest',
			plot_bgcolor: 'transparent',
			paper_bgcolor: 'transparent',
			height: 180,
			width: 411,
			yaxis: {
				title: { text: 'Requests' },
				gridcolor: 'gray',
				showgrid: false,
				fixedrange: true,
			},
			xaxis: {
				visible: false,
			},
			dragmode: false,
		};
	}

	const colors = [
		'#3FCF8E', // Green
		'#E46161', // Red
		'#EBEB81', // Yellow
	];

	function pieChart() {
		const deviceCount: ValueCount = {};
		for (let i = 0; i < data.length; i++) {
			const userAgent = getUserAgent(data[i][ColumnIndex.UserAgent]);
			const device = getDevice(userAgent);
			if (device in deviceCount) {
				deviceCount[device]++
			} else {
				deviceCount[device] = 1
			}
		}

		const devices = new Array(Object.keys(deviceCount).length);
		const counts = new Array(Object.keys(deviceCount).length);
		let i = 0;
		for (const [browser, count] of Object.entries(deviceCount)) {
			devices[i] = browser;
			counts[i] = count;
			i++;
		}
		return [
			{
				values: counts,
				labels: devices,
				type: 'pie',
				marker: {
					colors: colors,
				},
			},
		];
	}

	function devicePlotData() {
		return {
			data: pieChart(),
			layout: devicePlotLayout(),
			config: {
				responsive: true,
				showSendToCloud: false,
				displayModeBar: false,
			},
		};
	}

	function genPlot() {
		const plotData = devicePlotData();
		//@ts-ignore
		new Plotly.newPlot(
			plotDiv,
			plotData.data,
			plotData.layout,
			plotData.config,
		);
	}

	let plotDiv: HTMLDivElement;
	let mounted = false;
	onMount(() => {
		mounted = true;
	});

	$: data && mounted && genPlot();

	export let data: RequestsData, getUserAgent: (id: number) => string;
</script>

<div id="plotly">
	<div id="plotDiv" bind:this={plotDiv}>
		<!-- Plotly chart will be drawn inside this DIV -->
	</div>
</div>

<style>
	#plotDiv {
		margin-right: 20px;
	}
</style>
