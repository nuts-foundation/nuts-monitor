<template>
  <div>
    <div class="my-2">
      <div class="text-xl font-semibold text-gray-700 mb-2">Network</div>
      <div class="grid grid-cols-1 md:grid-cols-1 lg:grid-cols-3 gap-4">
        <div class="bg-white shadow rounded-lg p-4">
          <div class="font-semibold text-gray-700 mb-2">Last hour</div>
          <div id="viz_hour"></div>
        </div>
        <div class="bg-white shadow rounded-lg p-4">
          <div class="font-semibold text-gray-700 mb-2">Last day</div>
          <div id="viz_day"></div>
        </div>
        <div class="bg-white shadow rounded-lg p-4">
          <div class="font-semibold text-gray-700 mb-2">Last month</div>
          <div id="viz_month"></div>
        </div>
      </div>
      <hr class="border-gray-300 my-2">

    </div>
  </div>
</template>

<script>
import * as d3 from "d3";
export default {
  data () {
    return {
      'aggregatedTransactions': {
        'hourly': [],
        'daily': [],
        'monthly': []
      }
    }
  },
  mounted () {
    this.fetchData()
  },
  emits: ['statusUpdate'],
  watch: {},
  methods: {
    updateStatus (event) {
      this.$emit('statusUpdate', event)
    },
    fetchData () {
      this.feedbackMsg = ''

      this.$api.get('web/transactions/aggregated')
          .then(responseData => {
            this.aggregatedTransactions = responseData
            this.updateGraphs(responseData)
          })
          .catch(reason => {
            console.log('error while fetching data: ', reason)
          })
    },
    updateGraphs(responseData) {
      this.updateGraph('#viz_hour', responseData.hourly, '%H:%M')
      this.updateGraph('#viz_day', responseData.daily, '%H:%M')
      this.updateGraph('#viz_month', responseData.monthly,'%m-%d')
    },
    updateGraph(element, data, timeFormat) {
      console.log('updateGraph', element, data)

      // set the dimensions and margins of the graph
      const margin = {top: 20, right: 20, bottom: 30, left: 30},
          width = 600 - margin.left - margin.right,
          height = 300 - margin.top - margin.bottom;

      // Compute values.
      const X = d3.map(data, (d) => new Date(d.timestamp * 1000));
      const Y = d3.map(data, (d) => d.value);
      const Z = d3.map(data, (d) => d.contentType);

      const xDomain = d3.extent(X);
      let zDomain = Z;
      zDomain = new d3.InternSet(zDomain);

      // Omit any data not present in the z-domain.
      const I = d3.range(X.length).filter(i => zDomain.has(Z[i]));

      // Compute a nested array of series where each series is [[y1, y2], [y1, y2],
      // [y1, y2], …] representing the y-extent of each stacked rect. In addition,
      // each tuple has an i (index) property so that we can refer back to the
      // original data point (data[i]). This code assumes that there is only one
      // data point for a given unique x- and z-value.
      const series = d3.stack()
          .keys(zDomain)
          .value(([x, I], z) => Y[I.get(z)])
          (d3.rollup(I, ([i]) => i, i => X[i], i => Z[i]))
          .map(s => s.map(d => Object.assign(d, {i: d.data[1].get(s.key)})));

      // Compute the default y-domain. Note: diverging stacks can be negative.
      const yDomain = d3.extent(series.flat(2));

      // Construct scales and axes.
      const xRange = [margin.left, width - margin.right]
      const yRange = [height - margin.bottom, margin.top]
      const xScale = d3.scaleLinear(xDomain, xRange);
      const yScale = d3.scaleLinear(yDomain, yRange);
      const xFormat = d3.timeFormat(timeFormat);
      const color = d3.scaleOrdinal(zDomain, d3.schemeTableau10);
      const xAxis = d3.axisBottom(xScale).tickFormat(xFormat).tickSizeOuter(0);
      const yAxis = d3.axisLeft(yScale).ticks(height / 50);

      const area = d3.area()
          .x(({i}) => xScale(X[i]))
          .y0(([y1]) => yScale(y1))
          .y1(([, y2]) => yScale(y2));

      // clean up children of element first
      d3.select(element).selectAll("*").remove();

      const svg = d3.select(element)
          .append("svg")
          .attr("width", width)
          .attr("height", height)
          .attr("viewBox", [0, 0, width, height])
          .attr("style", "max-width: 100%; height: auto; height: intrinsic;");

      svg.append("g")
          .attr("transform", `translate(${margin.left},0)`)
          .call(yAxis)
          .call(g => g.select(".domain").remove())
          .call(g => g.selectAll(".tick line").clone()
              .attr("x2", width - margin.left - margin.right)
              .attr("stroke-opacity", 0.1))
          .call(g => g.append("text")
              .attr("x", -margin.left)
              .attr("y", 10)
              .attr("fill", "currentColor")
              .attr("text-anchor", "start")
              .text("count"));

      svg.append("g")
          .selectAll("path")
          .data(series)
          .join("path")
          .attr("fill", ([{i}]) => color(Z[i]))
          .attr("d", area)
          .append("title")
          .text(([{i}]) => Z[i]);

      svg.append("g")
          .attr("transform", `translate(0,${height - margin.bottom})`)
          .call(xAxis)
          .selectAll("text")
          .attr("transform", "rotate(-45),translate(-10,0)");
    }
  }
}
</script>
