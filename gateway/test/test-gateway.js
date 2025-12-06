import { ApolloClient, InMemoryCache, HttpLink, gql } from '@apollo/client'
import crossFetch from 'cross-fetch'
const client = new ApolloClient({
  link: new HttpLink({
    uri: 'http://localhost/graphql',
    fetch: crossFetch,
  }),
  cache: new InMemoryCache(),
});

// Test register
async function testRegister() {
  try {
    const result = await client.mutate({
      mutation: gql`
        mutation {
          register(
            email: "test@example.com"
            username: "testuser"
            password: "pass123"
          ) {
            token
            user {
              id
              email
              username
            }
          }
        }
      `,
    });
    console.log('✅ Register:', result.data);
    return result.data.register.token;
  } catch (error) {
    console.error('❌ Register failed:', error.message);
  }
}

// Test login
async function testLogin() {
  try {
    const result = await client.mutate({
      mutation: gql`
        mutation {
          login(
            email: "test@example.com"
            password: "pass123"
          ) {
            token
            user {
              id
              email
              username
            }
          }
        }
      `,
    });
    console.log('✅ Login:', result.data);
    return result.data.login.token;
  } catch (error) {
    console.error('❌ Login failed:', error.message);
  }
}

// Test get profile
async function testGetProfile(token) {
  try {
    const result = await client.query({
      query: gql`
        query {
          me {
            id
            email
            username
          }
        }
      `,
      context: {
        headers: {
          Authorization: `Bearer ${token}`,
        },
      },
    });
    console.log('✅ Get Profile:', result.data);
  } catch (error) {
    console.error('❌ Get Profile failed:', error.message);
  }
}

// Run tests
async function runTests() {
  console.log('🧪 Testing Gateway...\n');
  
  const token = await testRegister();
  console.log('\n---\n');
  
  await testLogin();
  console.log('\n---\n');
  
  if (token) {
    await testGetProfile(token);
  }
}

runTests();