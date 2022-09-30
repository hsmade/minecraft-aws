<template>
  <v-container>
    <v-card
        elevation="2"
        :loading="server.status==='UNKNOWN'"
    >
      <v-card-title>{{ server.name }}.${domain_name}</v-card-title>
      <v-card-subtitle>Status: {{ mapStatus(server.status) }}</v-card-subtitle>
      <v-card-text>
        <v-btn v-if="server.status === 'NONE'" @click="start_server()">Start</v-btn>
        <v-btn v-if="server.status !== 'NONE'" @click="stop_server()">Stop</v-btn>
        <v-alert>{{ error }}</v-alert>
        <v-list>
          <v-list-item v-for="(value,key) in server.tags" v-bind:key="key">
            <v-list-item-title>{{ key }}</v-list-item-title>
            <v-list-item-subtitle>{{ value }}</v-list-item-subtitle>
          </v-list-item>
        </v-list>
      </v-card-text>
    </v-card>
  </v-container>
</template>

<script>
  export default {
    name: 'ServerComponent',
    props: ["server"],
    data: () => ({
      error: ""
    }),
    methods: {
      start_server() {
        fetch("${server_stop}/?name="+this.server.name, { method: "PUT" })
            .then((response) => {
              if (!response.ok) {
                return response.text()
              }
            })
            .then((text) => {
              this.error = text
            })
      },
      stop_server() {
        fetch("${server_start}/?name="+this.server.name, { method: "DELETE" })
            .then((response) => {
              if (!response.ok) {
                return response.text()
              }
            })
            .then((text) => {
              this.error = text
            })
      },
      mapStatus(status) {
        switch(status) {
          case "NONE": return "Off"
          case "UNKNOWN": return "Starting"
          case "HEALTHY": return "On"
          default: return "Unknown"
        }
      }
    },
  }
</script>
