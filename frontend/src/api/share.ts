import { fetchURL, fetchJSON, removePrefix, createURL } from "./utils";

export async function list() {
  return fetchJSON<Share[]>("/api/shares");
}

export async function get(url: string) {
  url = removePrefix(url);
  return fetchJSON<Share>(`/api/share${url}`);
}

export async function remove(hash: string) {
  await fetchURL(`/api/share/${hash}`, {
    method: "DELETE",
  });
}

export async function create(
  url: string,
  password = "",
  expires = "",
  unit = "hours",
  deleteFiles = false,
  maxDownloads = 0
) {
  url = removePrefix(url);
  url = `/api/share${url}`;
  if (expires !== "") {
    url += `?expires=${expires}&unit=${unit}`;
  }
  let body = "{}";
  if (
    password != "" ||
    expires !== "" ||
    unit !== "hours" ||
    deleteFiles ||
    maxDownloads > 0
  ) {
    body = JSON.stringify({
      password: password,
      expires: expires.toString(), // backend expects string not number
      unit: unit,
      deleteFiles: deleteFiles,
      maxDownloads: maxDownloads,
    });
  }
  return fetchJSON(url, {
    method: "POST",
    body: body,
  });
}

export function getShareURL(share: Share) {
  return createURL("share/" + share.hash, {});
}
