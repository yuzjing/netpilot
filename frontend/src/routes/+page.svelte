<script>
  // 导入 onMount，这是一个特殊的Svelte函数，
  // 它会在组件被加载到页面上之后，立刻运行一次。
  // 这是进行初始数据获取的最佳位置。
  import { onMount } from 'svelte';

  // 1. 定义一个变量来存储从后端传来的消息。
  //    给它一个初始值，这样在数据加载完成前，页面上会显示这个。
  let message = '正在从Go后端加载数据...';

  // 2. onMount 会在页面准备好后，自动调用这个函数
  onMount(async () => {
    // 使用 try...catch 来优雅地处理可能发生的网络错误
    try {
      // 3. 使用浏览器内置的 fetch API，向我们的Go后端API地址发起请求。
      //    这就像在浏览器里隐形地打开了一个标签页去访问那个地址。
      const response = await fetch('http://localhost:8080/api/status');

      // 4. 将返回的响应解析为JSON格式。
      const data = await response.json();

      // 5. 【魔法时刻】从解析后的数据中，取出message字段，
      //    并用 '=' 赋给我们的message变量。
      //    Svelte会自动检测到这个变化，并立刻更新页面上所有用到message的地方！
      message = data.message;

    } catch (error) {
      // 如果Go后端没开，或者网络有问题，就会捕捉到错误。
      console.error('获取数据失败:', error);
      message = '无法连接到Go后端。请确认后端服务已启动！';
    }
  });
</script>

<main class="flex flex-col items-center justify-center min-h-screen bg-gray-100">
  
  <div class="p-8 bg-white rounded-lg shadow-md text-center">
    
    <h1 class="text-2xl font-bold text-gray-800 mb-4">
      NetPilot 前后端通信测试
    </h1>

    <p class="text-lg text-gray-600 border-t pt-4">
      <!-- 这里，{message} 会被Svelte自动替换成 <script> 标签里 message 变量的当前值 -->
      来自Go后端的消息：<span class="font-semibold text-blue-600">{message}</span>
    </p>

  </div>

</main>

<!-- 注意: 这里我们没有写 <style> 标签，因为我们用了Tailwind CSS! -->
<!-- 上面HTML里的 class="..." 就是Tailwind在起作用，帮我们美化页面。 -->