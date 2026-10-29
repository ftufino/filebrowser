import { fetchURL, removePrefix } from "./utils";

export async function send(url: string, to: string) {
  url = removePrefix(url);
  return fetchURL(`/api/email${url}`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ to }),
  });
}
