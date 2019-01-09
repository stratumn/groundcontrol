class Commit {
  constructor(
    public sha: string,
    public headline: string,
    public message: string,
    public author: string,
    public date: Date,
  ) {
  }

  public id(): string {
    return this.sha;
  }
}

export default Commit;
