// Invalidate asynchronous snapshots across task mutations and reject late progress events.
class TaskSnapshots {
  constructor(tasks) { this.tasks = tasks; this.epoch = 0; this.sequence = 0; this.applied = 0; }
  invalidate() { this.epoch++; }
  request() { return { epoch: this.epoch, sequence: ++this.sequence }; }
  apply(token, tasks) {
    if (token.epoch !== this.epoch || token.sequence < this.applied) return false;
    this.applied = token.sequence;
    this.tasks.clear();
    for (const task of tasks || []) this.tasks.set(task.id, task);
    return true;
  }
  progress(task) {
    if (!task?.id || !this.tasks.has(task.id)) return false;
    this.invalidate();
    this.tasks.set(task.id, task);
    return true;
  }
  clear() { this.invalidate(); this.tasks.clear(); }
  remove(id) { this.invalidate(); this.tasks.delete(id); }
}
globalThis.TaskSnapshots = TaskSnapshots;
