import config from "../../config.json";

export function getCdnAddress(id: string): string {
  return `${config.cdnUri}/${id}`;
}
