(() => {
  let page = 0;
  let haveErrored = false;
  let loading = false;
  let atBottom = false;

  const loadPosts = async () => {
    if (haveErrored || loading) {
      return;
    }
    loading = true;
    try {
      const res = await fetch('/api/posts?page=' + page);
      if (res.status !== 200) {
        throw new Error('bad response: ' + res.statusText);
      }
      const content = await res.json();
      appendContent(content.items);
      page++;
    } catch (err) {
      haveErrored = true;
      alert(err);
    } finally {
      loading = false;
    }
  };

  const appendContent = (content) => {
    const el = document.getElementById('content');
    content.forEach(post => {
      el.appendChild(makePost(post));
    });
  }

  const makePost = (post) => {
    const el = document.createElement('article');

    if (post.thumbnail) {
      const img = document.createElement('a');
      img.href = post.url;
      img.className = 'thumbnail';
      img.style.backgroundImage = `url(${post.thumbnail})`;
      img.target = '_blank';
      img.rel = 'noopener';
      el.appendChild(img);
    }

    const inner = document.createElement('div');
    inner.className = 'inner'
    el.appendChild(inner);

    const h2 = document.createElement('h2');
    inner.appendChild(h2);

    const a = document.createElement('a');
    a.href = post.url;
    a.innerHTML = post.title;
    a.target = '_blank';
    a.rel = 'noopener';
    h2.appendChild(a);

    const h3 = document.createElement('h3');
    h3.innerHTML = `${post.source} | ${new Date(post.timestamp).toLocaleString()}`;
    inner.appendChild(h3);

    const p = document.createElement('p');
    p.innerHTML = post.description;
    inner.appendChild(p);

    return el;
  }

  const onScroll = () => {
    if (atBottom) {
      return;
    }
    atBottom = (window.innerHeight + Math.round(window.scrollY)) >= document.body.offsetHeight;
    setTimeout(async () => {
      if ((window.innerHeight + Math.round(window.scrollY)) >= document.body.offsetHeight) {
        await loadPosts();
      }
      atBottom = false;
    }, 1000);
  }

  document.addEventListener('DOMContentLoaded', async () => {
    await loadPosts();
  })
  window.addEventListener('scroll', onScroll);
})();