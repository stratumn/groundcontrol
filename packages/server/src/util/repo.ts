export function ownerAndName(repo: string) {
  return repo.split(":")[1].split(".")[0].split("/").slice(0, 2);
}