import { ref } from "vue";
import * as email from "@/api/email";

// Whether the current user can send email (personal or global SMTP).
export const emailConfigured = ref(false);

export async function refreshEmailStatus() {
  try {
    emailConfigured.value = (await email.getSmtp()).configured;
  } catch {
    emailConfigured.value = false;
  }
}
