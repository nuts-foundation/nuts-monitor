<template>
  <div class="my-2">
    <div class="my-2 w-full h-full">
      <div class="text-xl font-semibold text-gray-700 mb-2">Network Topology</div>
      <svg ref="svg" class="w-full min-w-full" :height="svgHeight"></svg>
    </div>
  </div>
</template>

<script>
import * as d3 from "d3";
import tip from "d3-tip";

export default {
  data() {
    return {
      svgHeight: 0
    }
  },
  mounted () {
    this.fetchData()
  },
  methods: {
    updateStatus (event) {
      this.$emit('statusUpdate', event)
    },
    fetchData () {
      this.feedbackMsg = ''
      this.$api.get('web/network_topology')
          .then(responseData => {
            this.updateGraph(responseData)
          })
          .catch(reason => {
            console.log('error while fetching network topology: ', reason)
          })
    },
    setSVGHweight() {
      const svg = d3.select(this.$refs.svg);
      const bbox = svg.node().getBBox();
      if (bbox.height > 400) {
        this.svgHeight = bbox.height + 20;
      } else {
        this.svgHeight = 400
      }
    },
    updateGraph(data) {
      console.log(data)
      let edges = {}
      const vertices = data.vertices.map((peerID) => ({ id: peerID }));
      if (data.edges) {
        edges = data.edges.map(([from, to]) => ({
          source: from,
          target: to,
        }));
      }

      const svg = d3.select(this.$refs.svg);
      const simulation = d3
          .forceSimulation(vertices)
          .force("link", d3.forceLink(edges).id((d) => d.id).distance(200))
          .force("charge", d3.forceManyBody().strength(-1000))
          .force("center", d3.forceCenter(300, 200));
      const link = svg
          .selectAll(".link")
          .data(edges)
          .enter()
          .append("line")
          .attr("stroke", "black")
          .attr("stroke-width", 2)
          .attr("class", "link");

      // Define the tooltip
      const tooltip = tip()
          .attr("class", "d3-tip")
          .html( (d) => d.target.id)
          .direction('n')
          .offset([-3, 0])
      svg.call(tooltip)
      // vertices
      const circles = svg
          .selectAll(".node")
          .data(vertices)
          .enter()
          .append("circle")
          .attr("class", "node")
          .attr("r", 10)
          .attr("id", (d)=> d.id )
          .on('mouseover', tooltip.show)
          .on('mouseout', tooltip.hide)
          .call(
              d3
                  .drag()
                  .on("start", (event, d) => {
                    if (!event.active) {
                      simulation.alphaTarget(0.3).restart();
                    }
                    d.fx = d.x;
                    d.fy = d.y;
                  })
                  .on("drag", (event, d) => {
                    d.fx = event.x;
                    d.fy = event.y;
                    this.setSVGHweight();
                  })
                  .on("end", (event, d) => {
                    if (!event.active) {
                      simulation.alphaTarget(0);
                    }
                    d.fx = null;
                    d.fy = null;
                    this.setSVGHweight();
                  })
          )


      simulation.on("tick", () => {
        link.attr("x1", (d) => d.source.x)
            .attr("y1", (d) => d.source.y)
            .attr("x2", (d) => d.target.x)
            .attr("y2", (d) => d.target.y);
        circles.attr("cx", (d) => d.x).attr("cy", (d) => d.y);
      });

      this.setSVGHweight()
    }
  }
}
</script>
