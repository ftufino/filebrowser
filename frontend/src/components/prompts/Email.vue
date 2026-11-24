<template>
  <div class="card floating" id="email">
    <div class="card-title">
      <h2>{{ $t("buttons.sendEmail") }}</h2>
    </div>

    <div class="card-content">
      <p class="small">
        {{ $t("prompts.emailSendHelp", { max: maxAttachmentMB }) }}
      </p>
      <p>
        <label for="email-recipient">{{ $t("prompts.emailRecipient") }}</label>
        <input
          id="email-recipient"
          class="input input--block"
          type="email"
          v-model.trim="emailTo"
          autocomplete="email"
          @keyup.enter="sendEmail"
        />
      </p>
    </div>

    <div class="card-action">
      <button
        class="button button--flat button--grey"
        @click="closeHovers"
        :aria-label="$t('buttons.close')"
        :title="$t('buttons.close')"
      >
        {{ $t("buttons.close") }}
      </button>
      <button
        class="button button--flat button--blue"
        :disabled="!emailTo || sendingEmail"
        @click="sendEmail"
        :aria-label="$t('buttons.sendEmail')"
        :title="$t('buttons.sendEmail')"
      >
        {{ sendingEmail ? $t("prompts.sendingEmail") : $t("buttons.sendEmail") }}
      </button>
    </div>
  </div>
</template>

<script>
import { mapActions, mapState } from "pinia";
import { useFileStore } from "@/stores/file";
import { useLayoutStore } from "@/stores/layout";
import * as api from "@/api/index";
import { emailMaxAttachmentMB } from "@/utils/constants";

export default {
  name: "email",
  data() {
    return {
      emailTo: "",
      sendingEmail: false,
    };
  },
  inject: ["$showError", "$showSuccess"],
  computed: {
    ...mapState(useFileStore, ["req", "selected", "selectedCount"]),
    maxAttachmentMB() {
      return emailMaxAttachmentMB;
    },
  },
  methods: {
    ...mapActions(useLayoutStore, ["closeHovers"]),
    async sendEmail() {
      if (!this.emailTo || this.sendingEmail || this.selectedCount === 0) {
        return;
      }
      this.sendingEmail = true;
      try {
        const names = this.selected.map((i) => this.req.items[i].name);
        await api.email.send(this.$route.path, this.emailTo, names);
        this.$showSuccess(this.$t("success.emailSent"));
        this.closeHovers();
      } catch (e) {
        this.$showError(e);
      } finally {
        this.sendingEmail = false;
      }
    },
  },
};
</script>
