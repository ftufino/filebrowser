import { fetchURL, fetchJSON, removePrefix } from "./utils";

export interface PersonalSmtp {
  host: string;
  port: number;
  username: string;
  password?: string;
  from: string;
  hasPassword?: boolean;
}

export interface SmtpStatus {
  configured: boolean;
  personal: PersonalSmtp | null;
}

export async function getSmtp() {
  return fetchJSON<SmtpStatus>("/api/smtp", {});
}

export async function saveSmtp(smtp: PersonalSmtp) {
  return fetchURL("/api/smtp", {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(smtp),
  });
}

export async function send(url: string, to: string, files?: string[]) {
  url = removePrefix(url);
  return fetchURL(`/api/email${url}`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ to, files }),
  });
}
