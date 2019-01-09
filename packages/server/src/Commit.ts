class Commit {
  constructor(
    public sha: string,
    public headline: string,
    public message: string,
    public author: string,
    public date: Date
  ) {
  }

  id(): string {
    return this.sha
  }
}

export default Commit;
