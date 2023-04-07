<template>
  <div>
    <div class="my-2">
      <div class="text-xl font-semibold text-gray-700 mb-2">Network</div>
      <div class="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
        <div class="bg-white shadow rounded-lg p-4">
          <div class="font-semibold text-gray-700 mb-2">Transactions</div>
          <div class="text-2xl font-bold text-gray-900">{{diagnostics.network.state.transaction_count}}</div>
        </div>
        <div class="bg-white shadow rounded-lg p-4">
          <div class="font-semibold text-gray-700 mb-2">LC</div>
          <div class="text-2xl font-bold text-gray-900">{{diagnostics.network.state.dag_lc_high}}</div>
        </div>
        <div class="bg-white shadow rounded-lg p-4">
          <div class="font-semibold text-gray-700 mb-2">Connections</div>
          <div class="text-2xl font-bold text-gray-900">{{diagnostics.network.connections.connected_peers_count}}</div>
        </div>
        <div class="bg-white shadow rounded-lg p-4">
          <div class="font-semibold text-gray-700 mb-2">Failed events</div>
          <div class="text-2xl font-bold text-gray-900">{{diagnostics.network.state.failed_events}}</div>
        </div>
      </div>
      <div class="mt-2">
        <table class="table-auto border-collapse w-full">
          <tbody>
          <tr>
            <td class="text-gray-700 py-2 px-1 font-semibold block md:inline-block">PeerID:</td>
            <td class="text-gray-900 py-2 px-1 font-bold block md:inline-block overflow-hidden whitespace-nowrap text-ellipsis">{{diagnostics.network.connections.peer_id}}</td>
          </tr>
          <tr>
            <td class="text-gray-700 py-2 px-1 font-semibold block md:inline-block">XOR:</td>
            <td class="text-gray-900 py-2 px-1 font-bold block md:inline-block overflow-hidden whitespace-nowrap text-ellipsis">{{diagnostics.network.state.dag_xor}}</td>
          </tr>
          <tr>
            <td class="text-gray-700 py-2 px-1 font-semibold block md:inline-block">Database size:</td>
            <td class="text-gray-900 py-2 px-1 font-bold block md:inline-block">{{diagnostics.network.state.stored_database_size_bytes}}</td>
          </tr>
          </tbody>
        </table>
      </div>
      <hr class="border-gray-300 my-2">
      <div class="text-xl font-semibold text-gray-700 mb-2 mt-4">VDR</div>
      <div class="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4 mb-4">
        <div class="bg-white shadow rounded-lg p-4">
          <div class="font-semibold text-gray-700 mb-2">DID Document count</div>
          <div id="documents_count" class="text-2xl font-bold text-gray-900">{{diagnostics.vdr.did_documents_count}}</div>
        </div>
        <div class="bg-white shadow rounded-lg p-4">
          <div class="font-semibold text-gray-700 mb-2">Conflicted own</div>
          <div class="text-2xl font-bold text-gray-900">{{diagnostics.vdr.conflicted_did_documents.owned_count}}</div>
        </div>
        <div class="bg-white shadow rounded-lg p-4">
          <div class="font-semibold text-gray-700 mb-2">Conflicted total</div>
          <div class="text-2xl font-bold text-gray-900">{{diagnostics.vdr.conflicted_did_documents.total_count}}</div>
        </div>
      </div>
      <hr class="border-gray-300 my-2">
      <div class="text-xl font-semibold text-gray-700 mb-2 mt-4">VCR</div>
      <div class="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4 mb-4">
        <div class="bg-white shadow rounded-lg p-4">
          <div class="font-semibold text-gray-700 mb-2">Credentials</div>
          <div class="text-2xl font-bold text-gray-900">{{diagnostics.vcr.credential_count}}</div>
        </div>
        <div class="bg-white shadow rounded-lg p-4">
          <div class="font-semibold text-gray-700 mb-2">Verifier revocations</div>
          <div class="text-2xl font-bold text-gray-900">{{diagnostics.vcr.verifier.revocations_count}}</div>
        </div>
        <div class="bg-white shadow rounded-lg p-4">
          <div class="font-semibold text-gray-700 mb-2">Issued credentials</div>
          <div class="text-2xl font-bold text-gray-900">{{diagnostics.vcr.issuer.issued_credentials_count}}</div>
        </div>
        <div class="bg-white shadow rounded-lg p-4">
          <div class="font-semibold text-gray-700 mb-2">Issuer revocations</div>
          <div class="text-2xl font-bold text-gray-900">{{diagnostics.vcr.issuer.revoked_credentials_count}}</div>
        </div>
      </div>
      <hr class="border-gray-300 my-2">
      <div class="text-xl font-semibold text-gray-700 mb-2 mt-4">Status</div>
      <div class="mt-2">
        <table class="table-auto border-collapse w-full">
          <tbody>
          <tr>
            <td class="text-gray-700 py-2 px-1 font-semibold block md:inline-block">Software version:</td>
            <td class="text-gray-900 py-2 px-1 font-bold block md:inline-block">{{diagnostics.status.software_version}}</td>
          </tr>
          <tr>
            <td class="text-gray-700 py-2 px-1 font-semibold block md:inline-block">GIT commit hash:</td>
            <td class="text-gray-900 py-2 px-1 font-bold block md:inline-block overflow-hidden whitespace-nowrap text-ellipsis">{{diagnostics.status.git_commit}}</td>
          </tr>
          <tr>
            <td class="text-gray-700 py-2 px-1 font-semibold block md:inline-block">OS architecture:</td>
            <td class="text-gray-900 py-2 px-1 font-bold block md:inline-block">{{diagnostics.status.os_arch}}</td>
          </tr>
          <tr>
            <td class="text-gray-700 py-2 px-1 font-semibold block md:inline-block">Uptime:</td>
            <td class="text-gray-900 py-2 px-1 font-bold block md:inline-block">{{formatTime(diagnostics.status.uptime)}}</td>
          </tr>
          </tbody>
        </table>
      </div>
    </div>

    <router-view name="modal" @statusUpdate="updateStatus"></router-view>
  </div>
</template>

<script>
export default {
  data () {
    return {
      'diagnostics': {
        'network': {
          'state': {
            'transaction_count': -1,
            'dag_lc_high': -1,
            'failed_events': -1,
            'dag_xor': '0000000000000000000000000000000000000000000000000000000000000000',
            'stored_database_size_bytes': -1
          },
          'connections':{
            'connected_peers_count': -1,
            'peer_id': '0000000000000000000000000000000000000000000000000000000000000000'
          }
        },
        'status': {
          'git_commit': '0000000000000000000000000000000000000000000000000000000000000000',
          'os_arch': 'unknown',
          'software_version': 'master',
          'uptime': -1
        },
        'vdr': {
          'did_documents_count': -1,
          'conflicted_did_documents': {
            'owned_count': -1,
            'total_count': -1
          }
        },
        'vcr': {
          'credential_count': -1,
          'verifier': {
            'revocations_count': -1
          },
          'issuer': {
            'issued_credentials_count': -1,
            'revoked_credentials_count': -1
          }
        }
      }
    }
  },
  created () {
    // watch the params of the route to fetch the data again
    this.$watch(
        () => this.$route.params,
        () => {
          this.fetchData()
        },
        // fetch the data when the view is created and the data is
        // already being observed
        { immediate: true }
    )
  },
  emits: ['statusUpdate'],
  watch: {
    '$route.params': {
      handler (toParams, previousParams) {
        // Fetch data when the route change (e.g. from the modal back to the list)
        this.fetchData()
      },
      immediate: true
    }
  },
  methods: {
    updateStatus (event) {
      this.$emit('statusUpdate', event)
    },
    fetchData () {
      this.feedbackMsg = ''

      this.$api.get('web/diagnostics')
        .then(responseData => {
          this.diagnostics = responseData
        })
        .catch(reason => {
          console.log('error while fetching endpoints: ', reason)
        })
    },
    formatTime (nanoseconds) {
      if (nanoseconds < 60000000000) {
        return (nanoseconds / 1000000000) + ' s';
      } else if (nanoseconds < 3600000000000) {
        const minutes = Math.floor(nanoseconds / 60000000000);
        const seconds = (nanoseconds % 60000000000) / 1000000000;
        return `${minutes} min ${seconds} s`;
      } else if (nanoseconds < 86400000000000) {
        const hours = Math.floor(nanoseconds / 3600000000000);
        const minutes = Math.floor((nanoseconds % 3600000000000) / 60000000000);
        const seconds = (nanoseconds % 60000000000) / 1000000000;
        return `${hours} hr ${minutes} min ${seconds} s`;
      } else {
        const days = Math.floor(nanoseconds / 86400000000000);
        const hours = Math.floor((nanoseconds % 86400000000000) / 3600000000000);
        const minutes = Math.floor((nanoseconds % 3600000000000) / 60000000000);
        const seconds = (nanoseconds % 60000000000) / 1000000000;
        return `${days} days ${hours} hr ${minutes} min ${seconds} s`;
      }
    }
  }
}
</script>
