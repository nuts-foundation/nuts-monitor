<template>
  <div class="my-2">
    <div class="my-2 w-full">
      <div class="text-xl font-semibold text-gray-700 mb-2">Network Topology</div>
      <div id="container" class="h-full">
        <svg id="topology" ref="svg" class="network-svg z-10"></svg>
        <div id="tooltip" style="opacity: 0" class="absolute z-0 m-2 rounded-lg bg-white antialiased shadow-xl">
          <div id="tooltip-title" style="font-size: 1vw" class="flex shrink-0 items-center p-2 font-sans font-semibold leading-snug text-blue-gray-900 antialiased">
            Here the name of the node will appear, this will be the NutsComm address or PeerID.
          </div>
          <div id="tooltip-text" style="font-size: 0.8vw" class="relative border-t border-b border-t-blue-gray-100 border-b-blue-gray-100 p-2 font-sans font-light leading-relaxed text-blue-gray-500 antialiased">
            Here a list of stats will appear for the selected node.
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import * as d3 from "d3";

export default {
  data() {
    return {
      items: []
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
            this.updateList(responseData)
          })
          .catch(reason => {
            console.log('error while fetching network topology: ', reason)
          })
    },
    updateList(data) {
      data.peers.forEach(v => {
        const node = {name: v.peer_id}
        this.items.push(node)
      })
    },
    updateGraph(data) {
      console.log(data)
      const colorin = "#00f"
      const colornone = "#ccc"

      const line = d3.lineRadial()
          .curve(d3.curveBundle.beta(0.85))
          .radius(d => d.y)
          .angle(d => d.x)

      // sizing of viewport
      const svgElement = document.getElementById('topology');
      const rect = svgElement.getBoundingClientRect();
      const height = rect.height;
      const radius = rect.height / 2 * 9/10;

      const tree = d3.cluster()
          .size([2 * Math.PI, radius * 1/3])

      const root = tree(nodeLinks(d3.hierarchy(flatHierarchy(data))
          .sort((a, b) => d3.ascending(a.height, b.height) || d3.ascending(a.data.name, b.data.name))));

      const svg = d3.select(this.$refs.svg).attr("viewBox", [-height/2 * 4/5, -height/2, height, height]);
      const node = svg.append("g")
          .attr("font-family", "sans-serif")
          .style("font-size", "1vw")
          .selectAll("g")
          .data(root.leaves())
          .join("g")
          .attr("transform", d => `rotate(${d.x * 180 / Math.PI - 90}) translate(${d.y},0)`)
          .append("text")
          .attr("dy", "0.31em")
          .attr("x", d => d.x < Math.PI ? 6 : -6)
          .attr("text-anchor", d => d.x < Math.PI ? "start" : "end")
          .attr("transform", d => d.x >= Math.PI ? "rotate(180)" : null)
          .attr("fill", d => nodeColor(d))
          .text(d => d.data.name)
          .each(function(d) { d.text = this; })
          .on("mouseover", overed)
          .on("mouseout", outed);

      const link = svg.append("g")
          .attr("stroke", colornone)
          .attr("fill", "none")
          .selectAll("path")
          .data(root.leaves().flatMap(leaf => leaf.connections))
          .join("path")
          .style("mix-blend-mode", "multiply")
          .attr("d", ([i, o]) => line(i.path(o)))
          .each(function(d) { d.path = this; });

      function overed(event, d) {
        link.style("mix-blend-mode", null);
        d3.select(this).attr("font-weight", "bold");
        d3.selectAll(d.connections.map(d => d.path)).attr("stroke", colorin).raise();
        d3.selectAll(d.connections.map(([d]) => d.text)).attr("fill", colorin).attr("font-weight", "bold");

        d3.select('#tooltip')
            .style('opacity', 1)
            .style('left', event.pageX + 10 + 'px')
            .style('top', event.pageY + 10 + 'px');
        d3.select('#tooltip-title').text(d.data.name);
        d3.select('#tooltip-text').html(`
<ul>
<li><label for="peer_id">PeerID:</label><span id="peer_id">${d.data.raw.peer_id}</span></li>
<li><label for="peer_id">#Transactions:</label><span id="peer_id">${d.data.raw.tx_count}</span></li>
<li><label for="peer_id">#Connections:</label><span id="peer_id">${d.connections.length}</span></li>
<li><label for="peer_id">Address:</label><span id="peer_id">${d.data.raw.address}</span></li>
<li><label for="peer_id">NodeDID:</label><span id="peer_id">${d.data.raw.node_did}</span></li>
<li><label for="peer_id">Certificate:</label><span id="peer_id">${d.data.raw.cn}</span></li>
<li><label for="peer_id">Contact name:</label><span id="peer_id">${d.data.raw.contact_name}</span></li>
<li><label for="peer_id">Contact phone:</label><span id="peer_id">${d.data.raw.contact_phone}</span></li>
<li><label for="peer_id">Contact email:</label><span id="peer_id">${d.data.raw.contact_email}</span></li>
<li><label for="peer_id">Contact webaddress:</label><span id="peer_id">${d.data.raw.contact_web}</span></li>
<li><label for="peer_id">Software ID:</label><span id="peer_id">${d.data.raw.software_id}</span></li>
<li><label for="peer_id">Software version:</label><span id="peer_id">${d.data.raw.software_version}</span></li>
</ul>`);
      }

      function outed(event, d) {
        link.style("mix-blend-mode", "multiply");
        d3.select(this).attr("font-weight", null);
        d3.selectAll(d.connections.map(d => d.path)).attr("stroke", null);
        d3.selectAll(d.connections.map(([d]) => d.text)).attr("fill", nodeColor(d)).attr("font-weight", null);

        d3.select('#tooltip')
            .style('opacity', 0)
            .style('left', -1000 + 'px')
            .style('top', -1000 + 'px');
      }

      function nodeLinks(root) {
        const map = new Map(root.leaves().map(d => [d.data.raw.peer_id, d]));
        for (const d of root.leaves()) d.connections = d.data.connections.map(i => [d, map.get(i)]);
        return root;
      }

      function flatHierarchy(data) {
        const root = {name: 'root', children: []}
        const map = new Map()
        data.peers.forEach(v => {
          const peerName = name(v)
          const node = {
            name: peerName,
            raw: v,
            connections: []
          }
          map.set(v.peer_id, node)
          root.children.push(node)
        })
        data.edges.forEach(e => {
          map.get(e[0]).connections.push(e[1])
          map.get(e[1]).connections.push(e[0])
        })
        return root
      }

      function name(peer) {
        if (peer.address != "") {
          return peer.address
        }
        return peer.peer_id
      }

      function nodeColor(node) {
        return node.data.raw.authenticated ? "#7ebd7e" : "#ff6961";
      }

      return svg.node();
    }
  }

}
</script>
