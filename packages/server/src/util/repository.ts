export function ownerAndName(repository: string) {
  return repository.split(":")[1].split(".")[0].split("/").slice(0, 2);
}