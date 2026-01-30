import json
import requests
import time
from datetime import datetime

# Configuration
BASE_URL = "http://localhost:9000"
TEST_CASES_FILE = "error_span_test_cases.json"
OUTPUT_FILE = "llm_test_results.json"

# Simple question for all test cases
QUESTION = "Explain the root cause, impact, and suggested action for this error."


def load_test_cases():
    """Load test cases from JSON file"""
    with open(TEST_CASES_FILE, 'r') as f:
        data = json.load(f)
    return data['test_cases']


def register_user():
    """Register and get token"""
    print("Registering user...")
    response = requests.post(f"{BASE_URL}/register")
    response.raise_for_status()
    data = response.json()
    print(f"✓ Registered - User ID: {data['user_id']}")
    return data['token'], data['user_id']


def upload_trace(token, trace):
    """Upload trace and get upload_id"""
    headers = {
        "Authorization": f"Bearer {token}",
        "Content-Type": "application/json"
    }
    
    # Send the trace as-is (already in OTLP format)
    response = requests.post(
        f"{BASE_URL}/upload-trace",
        headers=headers,
        json=trace
    )
    response.raise_for_status()
    return response.json()['upload_id']

def query_llm(token, upload_id, span_id, question):
    """Query LLM for explanation"""
    headers = {
        "Authorization": f"Bearer {token}",
        "Content-Type": "application/json"
    }
    
    payload = {
        "upload_id": upload_id,
        "span_id": span_id,
        "question": question
    }
    
    response = requests.post(
        f"{BASE_URL}/explain-span",
        headers=headers,
        json=payload
    )
    response.raise_for_status()
    return response.json()


def run_tests():
    """Main test runner"""
    print(f"\n{'='*60}")
    print("Starting LLM Answer Collection")
    print(f"{'='*60}\n")
    
    # Load test cases
    test_cases = load_test_cases()
    print(f"Loaded {len(test_cases)} test cases\n")
    
    # Register user once
    token, user_id = register_user()
    
    # Results storage
    results = {
        "metadata": {
            "timestamp": datetime.now().isoformat(),
            "user_id": user_id,
            "total_tests": len(test_cases),
            "base_url": BASE_URL
        },
        "results": []
    }
    
    # Process each test case
    for idx, test_case in enumerate(test_cases, 1):
        print(f"\n[{idx}/{len(test_cases)}] Processing: {test_case['name']}")
        
        result = {
            "test_name": test_case['name'],
            "span_id": test_case['span_id'],
            "expected_answer": test_case['expected_answer'],
            "status": "pending"
        }
        
        try:
            # Upload trace
            print("  → Uploading trace...")
            upload_id = upload_trace(token, test_case['trace'])
            result['upload_id'] = upload_id
            print(f"  ✓ Upload ID: {upload_id}")
            
            # Query LLM (may take time)
            print(f"  → Querying LLM (this may take a few seconds)...")
            start_time = time.time()
            
            llm_response = query_llm(
                token,
                upload_id,
                test_case['span_id'],
                QUESTION
            )
            
            elapsed = time.time() - start_time
            
            result['llm_answer'] = llm_response.get('answer', llm_response)
            result['response_time_seconds'] = round(elapsed, 2)
            result['status'] = "success"
            
            print(f"  ✓ LLM responded in {elapsed:.2f}s")
            
        except requests.exceptions.HTTPError as e:
            result['status'] = "error"
            result['error'] = str(e)
            result['error_details'] = e.response.text if e.response else None
            print(f"  ✗ HTTP Error: {e}")
            
        except Exception as e:
            result['status'] = "error"
            result['error'] = str(e)
            print(f"  ✗ Error: {e}")
        
        results['results'].append(result)
        
        # Small delay between requests
        time.sleep(1)
    
    # Save results
    print(f"\n{'='*60}")
    print("Saving results...")
    with open(OUTPUT_FILE, 'w') as f:
        json.dump(results, f, indent=2)
    
    # Summary
    success_count = sum(1 for r in results['results'] if r['status'] == 'success')
    print(f"\n{'='*60}")
    print("Test Summary")
    print(f"{'='*60}")
    print(f"Total tests: {len(test_cases)}")
    print(f"Successful: {success_count}")
    print(f"Failed: {len(test_cases) - success_count}")
    print(f"\nResults saved to: {OUTPUT_FILE}")
    print(f"{'='*60}\n")


if __name__ == "__main__":
    try:
        run_tests()
    except KeyboardInterrupt:
        print("\n\nTest interrupted by user")
    except FileNotFoundError as e:
        print(f"\nError: Could not find {TEST_CASES_FILE}")
        print("Make sure the test cases file is in the same directory")
    except requests.exceptions.ConnectionError:
        print(f"\nError: Could not connect to {BASE_URL}")
        print("Make sure the API server is running")
    except Exception as e:
        print(f"\nUnexpected error: {e}")