<script>
  let count = 0;

  const increment = () => count++;
  const decrement = () => count--;
</script>

<div class="div-row">
  <button on:click={decrement}>Decrement</button>
  <p>{count}</p>
  <button on:click={increment}>Increment</button>
</div>

<style>
  .div-row {
    display: flex;
    flex-direction: row;
    gap: 8px;
    align-items: center;
  }

  .div-row button {
    height: 30px;
    background-color: #bbbbbb;
    border: none;
    border-radius: 8px;
  }

  .div-row button:hover {
    cursor: pointer;
    background-color: grey;
  }
</style>
